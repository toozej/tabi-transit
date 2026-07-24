package trimet

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestLiveArrivalsSmoke is an explicit credentialed compatibility check. It
// performs one bounded read-only request and never prints an AppID or response
// body. It is excluded from routine tests unless TRIMET_LIVE_SMOKE=1.
func TestLiveArrivalsSmoke(t *testing.T) {
	if os.Getenv("TRIMET_LIVE_SMOKE") != "1" {
		t.Skip("set TRIMET_LIVE_SMOKE=1 with a local TRIMET_APP_ID_FILE to run")
	}
	config, err := LoadConfig(os.Getenv, os.ReadFile)
	if err != nil {
		t.Fatalf("load live TriMet configuration: %v", err)
	}
	client, err := NewClient(config, nil, nil)
	if err != nil {
		t.Fatalf("construct live TriMet client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	arrivals, freshness, err := client.Arrivals(ctx, ArrivalsRequest{StopID: "8334", Minutes: 5})
	if err != nil {
		t.Fatalf("TriMet Arrivals V2 smoke request: %v", err)
	}
	if freshness.Source != SourceID || freshness.FetchedAt.IsZero() || !freshness.IsRealtime {
		t.Fatalf("unexpected arrivals freshness")
	}
	// A valid stop can legitimately have no imminent arrivals; response parsing
	// and safe freshness propagation are the assertions for this smoke check.
	for _, arrival := range arrivals {
		if arrival.StopID == "" || arrival.RouteID == "" {
			t.Fatalf("TriMet Arrivals V2 returned an invalid mapped arrival")
		}
	}
}
