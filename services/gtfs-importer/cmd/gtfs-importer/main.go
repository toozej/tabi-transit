package main

import (
	"context"
	"fmt"
	importer "github.com/toozej/tabi-transit/services/gtfs-importer"
	"os"
	"time"
)

func main() {
	s := importer.Service{Config: importer.Config{SourceID: os.Getenv("GTFS_SOURCE_ID"), Endpoint: os.Getenv("GTFS_ENDPOINT"), EndpointFile: os.Getenv("GTFS_ENDPOINT_FILE"), AllowedHosts: []string{os.Getenv("GTFS_ALLOWED_HOST")}, Timeout: 30 * time.Second, Policy: importer.DefaultPolicy()}}
	if err := s.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
