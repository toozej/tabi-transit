package main

import (
	"testing"
	"time"

	"github.com/toozej/tabi-transit/internal/application"
	"github.com/toozej/tabi-transit/internal/config"
)

func TestNewServerUsesConfiguredAddressAndBoundedTimeouts(t *testing.T) {
	t.Parallel()
	server := newServer(config.Config{ListenAddress: "127.0.0.1:8081", RateLimit: config.RateLimit{Requests: 1, Window: time.Second}}, application.Service{})
	if server.Addr != "127.0.0.1:8081" || server.Handler == nil {
		t.Fatalf("server = %#v", server)
	}
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 15*time.Second || server.WriteTimeout != 30*time.Second || server.IdleTimeout != time.Minute {
		t.Fatalf("timeouts = header:%s read:%s write:%s idle:%s", server.ReadHeaderTimeout, server.ReadTimeout, server.WriteTimeout, server.IdleTimeout)
	}
}
