package main

import (
	"github.com/toozej/tabi-transit/internal/api"
	"github.com/toozej/tabi-transit/internal/application"
	"github.com/toozej/tabi-transit/internal/config"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	c, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err.Error())
		os.Exit(1)
	}
	slog.Info("starting transit API", "address", c.ListenAddress)
	if err := http.ListenAndServe(c.ListenAddress, api.New(application.Service{}, c)); err != nil {
		slog.Error("API stopped", "error", err.Error())
		os.Exit(1)
	}
}
