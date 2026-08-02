package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toozej/tabi-transit/internal/api"
	"github.com/toozej/tabi-transit/internal/application"
	"github.com/toozej/tabi-transit/internal/config"
	"github.com/toozej/tabi-transit/internal/persistence"
	"github.com/toozej/tabi-transit/internal/persistence/sqlcgen"
	"github.com/toozej/tabi-transit/internal/sources/trimet"
)

func main() {
	os.Exit(run())
}

func run() int {
	c, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err.Error())
		return 1
	}
	service := application.Service{}
	var reader *persistence.PostgresReader
	var pool *pgxpool.Pool
	if c.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pool, err = pgxpool.New(ctx, c.DatabaseURL)
		if err == nil {
			err = pool.Ping(ctx)
		}
		cancel()
		if err != nil {
			if pool != nil {
				pool.Close()
			}
			slog.Error("database is unavailable", "error", err.Error())
			return 1
		}
		defer pool.Close()
		reader = persistence.NewPostgresReader(sqlcgen.New(pool))
		service = application.Service{
			Catalog:   application.PersistenceCatalog{Reader: reader},
			Vehicles:  application.PersistenceVehicleStore{Reader: reader},
			History:   application.PersistenceVehicleStore{Reader: reader},
			RiderInfo: application.PersistenceRiderInfo{Reader: reader},
		}
	}
	trimetConfig, err := trimet.LoadConfig(os.Getenv, os.ReadFile)
	if err != nil {
		slog.Error("invalid TriMet configuration", "error", err.Error())
		return 1
	}
	if trimetConfig.PlannerEnabled {
		client, clientErr := trimet.NewClient(trimetConfig, nil, nil)
		if clientErr != nil {
			slog.Error("invalid TriMet planner configuration", "error", clientErr.Error())
			return 1
		}
		service.Planning.Planner = application.NewTriMetPlanner(client)
	}
	options := []api.Option{}
	if pool != nil && reader != nil {
		options = append(options, api.WithReadiness(func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				return err
			}
			return reader.Ready(ctx)
		}))
	}
	server := newServer(c, service, options...)
	shutdown, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	slog.Info("starting transit API", "address", c.ListenAddress, "databaseConfigured", pool != nil)
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return 0
		}
		slog.Error("API stopped", "error", err.Error())
		return 1
	case <-shutdown.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("API shutdown failed", "error", err.Error())
			return 1
		}
	}
	return 0
}

func newServer(c config.Config, service application.Service, options ...api.Option) *http.Server {
	return &http.Server{
		Addr:              c.ListenAddress,
		Handler:           api.New(service, c, options...),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
