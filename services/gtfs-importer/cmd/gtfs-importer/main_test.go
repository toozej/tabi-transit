package main

import (
	"testing"
	"time"
)

func TestNewServiceBuildsBoundedImporterConfig(t *testing.T) {
	t.Parallel()
	values := map[string]string{
		"GTFS_SOURCE_ID":     "trimet",
		"GTFS_ENDPOINT":      "https://feeds.example/gtfs.zip",
		"GTFS_ENDPOINT_FILE": "/run/secrets/gtfs-endpoint",
		"GTFS_ALLOWED_HOST":  "feeds.example",
	}
	service := newService(func(name string) string { return values[name] })
	config := service.Config
	if config.SourceID != "trimet" || config.Endpoint != "https://feeds.example/gtfs.zip" || config.EndpointFile != "/run/secrets/gtfs-endpoint" || len(config.AllowedHosts) != 1 || config.AllowedHosts[0] != "feeds.example" {
		t.Fatalf("config = %#v", config)
	}
	if config.Timeout != 30*time.Second || config.Policy.MaxBytes == 0 {
		t.Fatalf("unsafe defaults: %#v", config)
	}
}
