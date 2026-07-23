package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toozej/tabi-transit/internal/api"
	"github.com/toozej/tabi-transit/internal/application"
	"github.com/toozej/tabi-transit/internal/config"
	"github.com/toozej/tabi-transit/internal/persistence"
	"github.com/toozej/tabi-transit/internal/persistence/sqlcgen"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	c, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err.Error())
		os.Exit(1)
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
			os.Exit(1)
		}
		defer pool.Close()
		reader = persistence.NewPostgresReader(sqlcgen.New(pool))
		service = application.Service{
			Catalog:   application.PersistenceCatalog{Reader: reader},
			Vehicles:  application.PersistenceVehicleStore{Reader: reader},
			RiderInfo: application.PersistenceRiderInfo{Reader: reader},
		}
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
	slog.Info("starting transit API", "address", c.ListenAddress, "databaseConfigured", pool != nil)
	if err := http.ListenAndServe(c.ListenAddress, api.New(service, c, options...)); err != nil {
		slog.Error("API stopped", "error", err.Error())
		os.Exit(1)
	}
}
