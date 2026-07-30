package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/toozej/tabi-transit/internal/api"
	"github.com/toozej/tabi-transit/internal/application"
	"github.com/toozej/tabi-transit/internal/config"
	"github.com/toozej/tabi-transit/internal/persistence"
)

// benchmarkVehicleHistory is intentionally deterministic: it represents the
// largest page the public endpoint accepts, rather than assuming a particular
// agency polling cadence or the total observations retained for a vehicle.
type benchmarkVehicleHistory struct {
	items []persistence.VehicleObservation
}

func (s benchmarkVehicleHistory) ListVehicleHistory(context.Context, persistence.VehicleHistoryFilter) ([]persistence.VehicleObservation, error) {
	return s.items, nil
}

func productionShapedVehicleHistory() []persistence.VehicleObservation {
	const pageSize = 500
	base := time.Date(2026, 7, 28, 18, 0, 0, 0, time.UTC)
	items := make([]persistence.VehicleObservation, 0, pageSize)
	for i := 0; i < pageSize; i++ {
		observedAt := base.Add(-time.Duration(i) * 30 * time.Second)
		item := persistence.VehicleObservation{
			VehicleID:       "trimet-rt:vehicle:2901",
			SourceID:        "trimet-rt",
			SourceVehicleID: "2901",
			Mode:            "bus",
			Coordinate: persistence.Coordinate{
				Longitude: -122.6765 + float64(i%19)*0.0001,
				Latitude:  45.5231 + float64(i%23)*0.0001,
			},
			ObservedAt: observedAt,
			FetchedAt:  observedAt.Add(2 * time.Second),
			Freshness:  persistence.FreshnessFresh,
		}
		// Real-time assignments can be absent. Exercise both response shapes
		// while making the fixture stable and representative of a full page.
		if i%5 != 0 {
			route := fmt.Sprintf("trimet-gtfs:route:%d", 4+i%48)
			item.RouteID = &route
		}
		if i%4 != 0 {
			trip := fmt.Sprintf("trimet-gtfs:trip:%d", 10000+i%300)
			item.TripID = &trip
		}
		items = append(items, item)
	}
	return items
}

func newVehicleHistoryBenchmarkServer() http.Handler {
	return api.New(application.Service{
		Catalog:  fakeCatalog{},
		Vehicles: fakeVehicles{},
		History:  benchmarkVehicleHistory{items: productionShapedVehicleHistory()},
	}, config.Config{RateLimit: config.RateLimit{Requests: 1_000_000, Window: time.Hour}})
}

// BenchmarkVehicleHistoryMaximumPage measures the local HTTP handler work for
// a 500-observation page: request parsing, response mapping, JSON encoding,
// ETag hashing, and writing to an in-memory recorder. It deliberately does not
// claim a PostgreSQL query-plan, connection-pool, TLS, or mobile-render result.
func BenchmarkVehicleHistoryMaximumPage(b *testing.B) {
	h := newVehicleHistoryBenchmarkServer()
	request := httptest.NewRequest(http.MethodGet, "/v1/vehicles/trimet-rt:vehicle:2901/history?from=2026-07-28T12:00:00Z&to=2026-07-28T18:00:00Z&limit=500", nil)
	request.RemoteAddr = "198.51.100.17:1234"
	sample := httptest.NewRecorder()
	h.ServeHTTP(sample, request)
	if sample.Code != http.StatusOK {
		b.Fatalf("sample response status = %d", sample.Code)
	}

	b.ReportAllocs()
	b.SetBytes(int64(sample.Body.Len()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := httptest.NewRecorder()
		h.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			b.Fatalf("response status = %d", response.Code)
		}
	}
}

func TestVehicleHistoryMaximumPageBenchmarkFixture(t *testing.T) {
	items := productionShapedVehicleHistory()
	if len(items) != 500 {
		t.Fatalf("benchmark observations = %d, want 500", len(items))
	}

	h := newVehicleHistoryBenchmarkServer()
	request := httptest.NewRequest(http.MethodGet, "/v1/vehicles/trimet-rt:vehicle:2901/history?from=2026-07-28T12:00:00Z&to=2026-07-28T18:00:00Z&limit=500", nil)
	request.RemoteAddr = "198.51.100.17:1234"
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("response status = %d: %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("ETag"); got == "" {
		t.Fatal("maximum page response lacks an ETag")
	}
	if response.Body.Len() == 0 {
		t.Fatal("maximum page response is empty")
	}
}
