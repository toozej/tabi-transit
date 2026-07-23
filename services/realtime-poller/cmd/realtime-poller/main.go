package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toozej/tabi-transit/internal/config"
	"github.com/toozej/tabi-transit/internal/persistence"
	poller "github.com/toozej/tabi-transit/services/realtime-poller"
)

func main() {
	pollConfig, err := poller.DefaultConfig()
	if errors.Is(err, poller.ErrDisabled) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	// This composition root is the only poller code that reads the database
	// secret. It never logs the URL and does not inspect local .env files.
	databaseURL, err := config.Secret("TABI_DATABASE_URL")
	if err != nil || databaseURL == "" {
		fmt.Fprintln(os.Stderr, "TABI_DATABASE_URL or TABI_DATABASE_URL_FILE is required when a realtime endpoint is configured")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	pool, err := pgxpool.New(connectCtx, databaseURL)
	if err == nil {
		err = pool.Ping(connectCtx)
	}
	cancel()
	if err != nil {
		if pool != nil {
			pool.Close()
		}
		slog.Error("realtime poller database unavailable", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Close()
	service := poller.Service{Config: pollConfig, Store: persistence.PostgresRealtimeWriter{DB: pool}}
	errCh := make(chan error, 2)
	if pollConfig.Endpoint != "" {
		go func() { errCh <- service.RunLoop(ctx) }()
	}
	if pollConfig.TripUpdatesEndpoint != "" {
		go func() { errCh <- service.RunTripUpdatesLoop(ctx) }()
	}
	slog.Info("starting realtime poller", "vehiclePositionsConfigured", pollConfig.Endpoint != "", "tripUpdatesConfigured", pollConfig.TripUpdatesEndpoint != "")
	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("realtime poller stopped", "error", err.Error())
		os.Exit(1)
	}
}
