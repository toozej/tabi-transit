package main

import (
	"context"
	"fmt"
	"os"
	"time"

	importer "github.com/toozej/tabi-transit/services/gtfs-importer"
)

func main() {
	s := newService(os.Getenv)
	if err := s.Run(context.Background()); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newService(getenv func(string) string) importer.Service {
	return importer.Service{Config: importer.Config{SourceID: getenv("GTFS_SOURCE_ID"), Endpoint: getenv("GTFS_ENDPOINT"), EndpointFile: getenv("GTFS_ENDPOINT_FILE"), AllowedHosts: []string{getenv("GTFS_ALLOWED_HOST")}, Timeout: 30 * time.Second, Policy: importer.DefaultPolicy()}}
}
