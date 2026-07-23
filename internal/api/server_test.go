package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/toozej/tabi-transit/internal/api"
	"github.com/toozej/tabi-transit/internal/application"
	"github.com/toozej/tabi-transit/internal/config"
	"github.com/toozej/tabi-transit/internal/persistence"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeVehicles struct {
	items []persistence.Vehicle
	err   error
}

func (f fakeVehicles) ListCurrentVehicles(context.Context, persistence.VehicleFilter) ([]persistence.Vehicle, error) {
	return f.items, f.err
}

type fakeCatalog struct{}

func (fakeCatalog) ListRoutes(context.Context, application.RouteQuery) (application.Page[application.Route], error) {
	return application.Page[application.Route]{Items: []application.Route{{ID: "fixture:route:20", Mode: "bus", ShortName: "20", LongName: "Fixture"}}, StaticFeedVersion: "fixture-v1"}, nil
}
func (fakeCatalog) ListStops(context.Context, application.StopQuery) (application.Page[application.Stop], error) {
	return application.Page[application.Stop]{Items: []application.Stop{{ID: "fixture:stop:1", Name: "Fixture Stop", Coordinate: persistence.Coordinate{Longitude: -122.67, Latitude: 45.52}, Modes: []string{"bus"}}}}, nil
}
func testServer(t *testing.T, v application.VehicleStore) http.Handler {
	t.Helper()
	c := config.Config{API: config.PublicAPI{Version: "0.1.0", MinimumAppVersion: "0.1.0", StaleThresholdSeconds: 90, StaticFeedVersion: "fixture-v1", StaticFeedPublishedAt: time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)}, RateLimit: config.RateLimit{Requests: 20, Window: time.Hour}}
	return api.New(application.Service{Catalog: fakeCatalog{}, Vehicles: v}, c)
}
func vehicle() persistence.Vehicle {
	at := time.Date(2026, 7, 22, 16, 30, 2, 0, time.UTC)
	route := "fixture:route:20"
	return persistence.Vehicle{ID: "fixture:vehicle:2901", SourceID: "fixture-rt", SourceVehicleID: "2901", RouteID: &route, Mode: "bus", Coordinate: persistence.Coordinate{Longitude: -122.67, Latitude: 45.52}, FetchedAt: at, ProcessedAt: at, Freshness: persistence.FreshnessFresh, SnapshotID: 42}
}
func request(h http.Handler, path string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.RemoteAddr = "198.51.100.7:123"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestVehicleEndpointsConformToCoreContract(t *testing.T) {
	h := testServer(t, fakeVehicles{items: []persistence.Vehicle{vehicle()}})
	w := request(h, "/v1/vehicles")
	if w.Code != 200 || w.Header().Get("ETag") == "" || w.Header().Get("X-Request-Id") == "" {
		t.Fatalf("unexpected response: %d %#v", w.Code, w.Header())
	}
	var got struct {
		SnapshotID string `json:"snapshotId"`
		Vehicles   []struct {
			ID        string         `json:"id"`
			Freshness map[string]any `json:"freshness"`
		} `json:"vehicles"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.SnapshotID != "42" || len(got.Vehicles) != 1 || got.Vehicles[0].ID != "fixture:vehicle:2901" || got.Vehicles[0].Freshness["status"] != "fresh" {
		t.Fatalf("contract response mismatch: %#v", got)
	}
	w = request(h, "/v1/vehicles/search?q=2901")
	if w.Code != 200 {
		t.Fatalf("search: %d %s", w.Code, w.Body.String())
	}
	w = request(h, "/v1/vehicles/fixture:vehicle:2901")
	if w.Code != 200 {
		t.Fatalf("detail: %d %s", w.Code, w.Body.String())
	}
}
func TestETagValidationAndBadInput(t *testing.T) {
	h := testServer(t, fakeVehicles{items: []persistence.Vehicle{vehicle()}})
	first := request(h, "/v1/config")
	r := httptest.NewRequest(http.MethodGet, "/v1/config", nil)
	r.Header.Set("If-None-Match", first.Header().Get("ETag"))
	r.RemoteAddr = "203.0.113.1:1"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNotModified {
		t.Fatalf("etag code %d", w.Code)
	}
	w = request(h, "/v1/vehicles?format=bad")
	if w.Code != 400 {
		t.Fatalf("bad format: %d", w.Code)
	}
	w = request(h, "/v1/stops")
	if w.Code != 400 {
		t.Fatalf("missing query: %d", w.Code)
	}
	w = request(h, "/v1/vehicles/not-an-id")
	if w.Code != 400 {
		t.Fatalf("bad ID: %d", w.Code)
	}
}
func TestUnavailableAndRateLimit(t *testing.T) {
	h := testServer(t, fakeVehicles{err: errors.New("database unavailable")})
	w := request(h, "/v1/vehicles")
	if w.Code != 503 || w.Header().Get("Retry-After") != "30" {
		t.Fatalf("unavailable %d %#v", w.Code, w.Header())
	}
	h = api.New(application.Service{Catalog: fakeCatalog{}, Vehicles: fakeVehicles{}}, config.Config{RateLimit: config.RateLimit{Requests: 2, Window: time.Hour}})
	if request(h, "/health/live").Code != 200 || request(h, "/health/live").Code != 200 || request(h, "/health/live").Code != 429 {
		t.Fatal("rate limit did not apply")
	}
}
