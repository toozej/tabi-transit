package main

import (
	"context"
	"errors"
	"testing"
	"time"

	poller "github.com/toozej/tabi-transit/services/realtime-poller"
)

func TestStartLoopsStartsOnlyConfiguredFeeds(t *testing.T) {
	t.Parallel()
	vehicleDone := errors.New("vehicle loop stopped")
	tripDone := errors.New("trip loop stopped")
	calledVehicle, calledTrip := false, false
	errs := startLoops(context.Background(), poller.Config{Endpoint: "https://example.invalid/vehicles", TripUpdatesEndpoint: "https://example.invalid/trips"}, func(context.Context) error {
		calledVehicle = true
		return vehicleDone
	}, func(context.Context) error {
		calledTrip = true
		return tripDone
	})
	seen := map[error]bool{}
	for range 2 {
		select {
		case err := <-errs:
			seen[err] = true
		case <-time.After(time.Second):
			t.Fatal("configured loops did not start")
		}
	}
	if !calledVehicle || !calledTrip || !seen[vehicleDone] || !seen[tripDone] {
		t.Fatalf("called vehicle=%v trip=%v, errors=%v", calledVehicle, calledTrip, seen)
	}
}
