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
	"strings"
	"testing"
	"time"
)

type fakeVehicles struct {
	items []persistence.Vehicle
	err   error
}

type fakeJourneyPlanner struct {
	request application.JourneyRequest
	plan    application.JourneyPlan
}

func (f *fakeJourneyPlanner) Plan(_ context.Context, request application.JourneyRequest) (application.JourneyPlan, error) {
	f.request = request
	return f.plan, nil
}

type fakeVehicleHistory struct {
	items []persistence.VehicleObservation
	err   error
	query persistence.VehicleHistoryFilter
}

func (f *fakeVehicleHistory) ListVehicleHistory(_ context.Context, q persistence.VehicleHistoryFilter) ([]persistence.VehicleObservation, error) {
	f.query = q
	return f.items, f.err
}

func (f fakeVehicles) ListCurrentVehicles(context.Context, persistence.VehicleFilter) ([]persistence.Vehicle, error) {
	return f.items, f.err
}

type fakeCatalog struct{}

type fakeRiderInfo struct {
	nearby []persistence.NearbyStop
	query  application.NearbyQuery
}

func (f *fakeRiderInfo) NearbyStops(_ context.Context, q application.NearbyQuery) ([]persistence.NearbyStop, error) {
	f.query = q
	return f.nearby, nil
}
func (f *fakeRiderInfo) Stop(context.Context, string) (persistence.StopDetail, error) {
	accessible := true
	return persistence.StopDetail{Stop: persistence.CatalogStop{ID: "fixture:stop:1", Name: "Fixture Stop", Coordinate: persistence.Coordinate{Longitude: -122.67, Latitude: 45.52}, Modes: []string{"bus"}, WheelchairAccessible: &accessible}, StaticFeedVersion: "fixture-v1", Freshness: persistence.StaticFreshness{Source: "normalized-static-gtfs", ActivatedAt: time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)}}, nil
}
func (f *fakeRiderInfo) Route(context.Context, string) (persistence.RouteDetail, error) {
	return persistence.RouteDetail{Route: persistence.CatalogRoute{ID: "fixture:route:20", Mode: "bus", ShortName: "20", LongName: "Fixture"}, Directions: []persistence.RouteDirection{{ID: 0, Headsign: "Downtown"}}, StaticFeedVersion: "fixture-v1", Freshness: persistence.StaticFreshness{ActivatedAt: time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)}}, nil
}
func (f *fakeRiderInfo) RouteStops(context.Context, string, *int) ([]persistence.RouteStop, string, error) {
	return []persistence.RouteStop{{Stop: persistence.CatalogStop{ID: "fixture:stop:1", Name: "Fixture Stop", Coordinate: persistence.Coordinate{Longitude: -122.67, Latitude: 45.52}, Modes: []string{"bus"}}, Sequence: 1}}, "fixture-v1", nil
}
func (f *fakeRiderInfo) RouteShapes(context.Context, string, *int) ([]persistence.RouteShape, string, error) {
	return []persistence.RouteShape{{ID: "fixture:shape:20", RouteID: "fixture:route:20", Coordinates: [][]float64{{-122.67, 45.52}, {-122.66, 45.53}}}}, "fixture-v1", nil
}
func (f *fakeRiderInfo) StopSchedule(_ context.Context, q persistence.ScheduleFilter) ([]persistence.ScheduleTime, string, string, error) {
	direction := 0
	return []persistence.ScheduleTime{{TripID: "fixture:trip:night", RouteID: "fixture:route:20", StopID: q.StopID, ServiceDate: q.ServiceDate, DirectionID: &direction, Headsign: "Night", ServiceDaySeconds: 90930, DepartureAt: time.Date(2026, 7, 23, 1, 15, 30, 0, time.UTC)}}, "fixture-v1", "", nil
}
func (f *fakeRiderInfo) StopArrivals(_ context.Context, q persistence.ArrivalFilter) ([]persistence.Arrival, error) {
	trip := "fixture:trip:night"
	sequence := 2
	return []persistence.Arrival{{ID: "fixture:arrival:night", StopID: q.StopID, RouteID: "fixture:route:20", Status: "scheduled", ScheduledAt: time.Date(2026, 7, 23, 1, 15, 30, 0, time.UTC), TripID: &trip, StopSequence: &sequence, Freshness: persistence.StaticFreshness{ActivatedAt: time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)}}}, nil
}
func (f *fakeRiderInfo) Alerts(context.Context, persistence.AlertFilter) ([]persistence.Alert, string, error) {
	return []persistence.Alert{{ID: "fixture:alert:one", Revision: "fixture-1", Header: "Fixture alert", Effect: "other", Source: "fixture", FetchedAt: time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC), ProcessedAt: time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)}}, "", nil
}
func (f *fakeRiderInfo) Alert(context.Context, string) (persistence.Alert, error) {
	return persistence.Alert{ID: "fixture:alert:one", Revision: "fixture-1", Header: "Fixture alert", Effect: "other", Source: "fixture", FetchedAt: time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC), ProcessedAt: time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)}, nil
}

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
func testRiderServer(t *testing.T, rider *fakeRiderInfo) http.Handler {
	t.Helper()
	c := config.Config{API: config.PublicAPI{Version: "0.1.0", MinimumAppVersion: "0.1.0", StaticFeedVersion: "fixture-v1", StaticFeedPublishedAt: time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)}, RateLimit: config.RateLimit{Requests: 50, Window: time.Hour}}
	return api.New(application.Service{Catalog: fakeCatalog{}, Vehicles: fakeVehicles{items: []persistence.Vehicle{vehicle()}}, RiderInfo: rider}, c)
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

func TestRequestIDOnlyPropagatesSafeCorrelationTokens(t *testing.T) {
	h := testServer(t, fakeVehicles{items: []persistence.Vehicle{vehicle()}})
	valid := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	valid.RemoteAddr = "198.51.100.7:123"
	valid.Header.Set("X-Request-Id", "mobile.release_42-abc")
	validResponse := httptest.NewRecorder()
	h.ServeHTTP(validResponse, valid)
	if got := validResponse.Header().Get("X-Request-Id"); got != "mobile.release_42-abc" || !strings.Contains(validResponse.Body.String(), `"requestId":"mobile.release_42-abc"`) {
		t.Fatalf("valid request ID was not propagated: header=%q body=%s", got, validResponse.Body.String())
	}

	for _, id := range []string{"mobile session", "trace/42", "<script>", strings.Repeat("a", 129)} {
		r := httptest.NewRequest(http.MethodGet, "/health/live", nil)
		r.RemoteAddr = "198.51.100.7:123"
		r.Header.Set("X-Request-Id", id)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		got := w.Header().Get("X-Request-Id")
		if got == id || !validRequestIDForTest(got) || !strings.Contains(w.Body.String(), `"requestId":"`+got+`"`) {
			t.Fatalf("unsafe request ID %q produced header=%q body=%s", id, got, w.Body.String())
		}
	}
}

func validRequestIDForTest(id string) bool {
	if !strings.HasPrefix(id, "req_") || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("._-", r)) {
			return false
		}
	}
	return true
}

func TestVehicleHistoryIsBoundedAndPaginated(t *testing.T) {
	at := time.Date(2026, 7, 28, 16, 30, 2, 0, time.UTC)
	route := "fixture:route:20"
	history := &fakeVehicleHistory{items: []persistence.VehicleObservation{{VehicleID: "fixture:vehicle:2901", SourceID: "fixture-rt", SourceVehicleID: "2901", RouteID: &route, Mode: "bus", Coordinate: persistence.Coordinate{Longitude: -122.67, Latitude: 45.52}, ObservedAt: at, FetchedAt: at, Freshness: persistence.FreshnessFresh}}}
	c := config.Config{API: config.PublicAPI{Version: "0.1.0"}, RateLimit: config.RateLimit{Requests: 20, Window: time.Hour}}
	h := api.New(application.Service{Catalog: fakeCatalog{}, Vehicles: fakeVehicles{}, History: history}, c)
	w := request(h, "/v1/vehicles/fixture:vehicle:2901/history?from=2026-07-28T15:00:00Z&to=2026-07-28T17:00:00Z&limit=1")
	if w.Code != http.StatusOK || w.Header().Get("ETag") == "" || !strings.Contains(w.Body.String(), `"retentionDays":30`) || !strings.Contains(w.Body.String(), `"nextCursor"`) {
		t.Fatalf("history response = %d %s", w.Code, w.Body.String())
	}
	if history.query.VehicleID != "fixture:vehicle:2901" || history.query.Limit != 1 || !history.query.From.Equal(time.Date(2026, 7, 28, 15, 0, 0, 0, time.UTC)) {
		t.Fatalf("history query = %#v", history.query)
	}
	for _, path := range []string{
		"/v1/vehicles/fixture:route:20/history",
		"/v1/vehicles/fixture:vehicle:2901/history?from=2026-06-01T00:00:00Z&to=2026-07-28T00:00:01Z",
		"/v1/vehicles/fixture:vehicle:2901/history?limit=501",
	} {
		if got := request(h, path); got.Code != http.StatusBadRequest {
			t.Fatalf("%s = %d", path, got.Code)
		}
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
func TestVehicleBoundingBoxFilterAndValidation(t *testing.T) {
	inside := vehicle()
	outside := vehicle()
	outside.ID = "fixture:vehicle:2902"
	outside.SourceVehicleID = "2902"
	outside.Coordinate = persistence.Coordinate{Longitude: -122.40, Latitude: 45.80}
	h := testServer(t, fakeVehicles{items: []persistence.Vehicle{inside, outside}})

	w := request(h, "/v1/vehicles?bbox=-122.68,45.51,-122.66,45.53")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"fixture:vehicle:2901"`) || strings.Contains(w.Body.String(), `"fixture:vehicle:2902"`) {
		t.Fatalf("bbox filter response = %d %s", w.Code, w.Body.String())
	}

	for _, path := range []string{
		"/v1/vehicles?bbox=-122.68,45.51,-122.66",
		"/v1/vehicles?bbox=-181,45.51,-122.66,45.53",
		"/v1/vehicles?bbox=-122.66,45.53,-122.68,45.51",
	} {
		if got := request(h, path); got.Code != http.StatusBadRequest {
			t.Fatalf("%s = %d, want 400", path, got.Code)
		}
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

func TestPhaseThreeFeatureGatesFailClosed(t *testing.T) {
	h := testServer(t, fakeVehicles{})
	config := request(h, "/v1/config")
	var configResponse struct {
		Features map[string]struct {
			Enabled bool   `json:"enabled"`
			Reason  string `json:"reason"`
		} `json:"features"`
		Sources map[string]struct {
			Enabled bool `json:"enabled"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(config.Body.Bytes(), &configResponse); err != nil || configResponse.Features["placeSearch"].Enabled || configResponse.Features["journeyPlanner"].Enabled || configResponse.Features["placeSearch"].Reason != "external_provider_gate_pending" || !configResponse.Sources["trimetStreetcar"].Enabled {
		t.Fatalf("feature config = %#v, err=%v", configResponse.Features, err)
	}
	if _, ok := configResponse.Sources["roseCityTransit"]; ok {
		t.Fatalf("Rose City must not be exposed as a separate source: %#v", configResponse.Sources)
	}
	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/v1/search?q=private+address"},
		{method: http.MethodGet, path: "/v1/geocode/reverse?lat=45.52&lon=-122.67"},
		{method: http.MethodPost, path: "/v1/journeys/plan"},
	} {
		r := httptest.NewRequest(tc.method, tc.path, nil)
		r.RemoteAddr = "198.51.100.7:123"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusServiceUnavailable || w.Header().Get("Retry-After") != "3600" {
			t.Fatalf("%s %s = %d, headers=%#v", tc.method, tc.path, w.Code, w.Header())
		}
		var response struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response.Error.Code != "feature_unavailable" {
			t.Fatalf("%s %s response = %s, err=%v", tc.method, tc.path, w.Body.String(), err)
		}
	}
}

func TestJourneyPlannerParsesNormalizedRequestAndReturnsPlan(t *testing.T) {
	depart := time.Date(2026, 7, 28, 18, 30, 0, 0, time.UTC)
	planner := &fakeJourneyPlanner{plan: application.JourneyPlan{PlanID: "fixture-plan", Provider: "trimet-web-services", Freshness: application.PlannerFreshness{Source: "trimet-web-services", FetchedAt: depart, ProcessedAt: depart, IsRealtime: true}, Itineraries: []application.Itinerary{{ID: "fixture-plan:1", DurationSeconds: 600, Transfers: 0, Legs: []application.JourneyLeg{{Mode: "bus", RouteID: "20"}}}}}}
	c := config.Config{API: config.PublicAPI{Version: "0.1.0", MinimumAppVersion: "0.1.0"}, RateLimit: config.RateLimit{Requests: 20, Window: time.Hour}}
	h := api.New(application.Service{Planning: application.PlanningFeatures{Planner: planner}}, c)
	body := `{"origin":{"type":"coordinate","coordinate":[-122.67,45.52]},"destination":{"type":"stop","stopId":"trimet:stop:8334"},"time":{"mode":"depart_at","value":"2026-07-28T18:30:00Z"},"preferences":{"modes":["bus"],"maxTransfers":1}}`
	r := httptest.NewRequest(http.MethodPost, "/v1/journeys/plan", strings.NewReader(body))
	r.RemoteAddr = "198.51.100.7:123"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"planId":"fixture-plan"`) || !strings.Contains(w.Body.String(), `"source":"trimet-web-services"`) {
		t.Fatalf("response = %d %s", w.Code, w.Body.String())
	}
	if planner.request.Origin.Coordinate == nil || planner.request.Origin.Coordinate.Longitude != -122.67 || planner.request.Destination.ID != "trimet:stop:8334" || planner.request.Time == nil || !planner.request.Time.Value.Equal(depart) {
		t.Fatalf("planner request = %#v", planner.request)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/v1/journeys/plan", strings.NewReader(`{"origin":{}}`))
	r.RemoteAddr = "198.51.100.7:123"
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid body = %d %s", w.Code, w.Body.String())
	}
}

func TestNotificationEndpointsFailClosedWithoutTokenExposure(t *testing.T) {
	h := testServer(t, fakeVehicles{})
	for _, tc := range []struct{ method, path, body string }{
		{http.MethodPost, "/v1/installations", `{"platform":"ios","locale":"en-US","timeZone":"America/Los_Angeles","appVersion":"1"}`},
		{http.MethodPut, "/v1/installations/fixture:installation:one/push-token", `{"platform":"ios","token":"ExponentPushToken[private-token]"}`},
		{http.MethodGet, "/v1/subscriptions", ""},
		{http.MethodPost, "/v1/subscriptions", `{"type":"service_alert"}`},
		{http.MethodDelete, "/v1/subscriptions/fixture:subscription:one", ""},
	} {
		r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		r.Header.Set("X-Tabi-Installation-Credential", "0123456789abcdef0123456789abcdef")
		r.RemoteAddr = "198.51.100.7:123"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusServiceUnavailable || w.Header().Get("Retry-After") != "3600" || !strings.Contains(w.Body.String(), `"feature_unavailable"`) || strings.Contains(w.Body.String(), "private-token") {
			t.Fatalf("%s %s = %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
}

func TestRiderInformationStaticResponsesAndNearbyGrouping(t *testing.T) {
	rider := &fakeRiderInfo{nearby: []persistence.NearbyStop{{ID: "fixture:stop:1", Name: "Bus", Mode: "bus", Coordinate: persistence.Coordinate{Longitude: -122.67, Latitude: 45.52}, DistanceMeters: 12}, {ID: "fixture:stop:2", Name: "Rail", Mode: "light_rail", Coordinate: persistence.Coordinate{Longitude: -122.68, Latitude: 45.53}, DistanceMeters: 40}}}
	h := testRiderServer(t, rider)
	w := request(h, "/v1/stops/nearby?latitude=45.52&longitude=-122.67&limit=10&limitPerMode=2&modes=bus,light_rail&wheelchairAccessible=true")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"distanceType":"straight_line"`) || !strings.Contains(w.Body.String(), `"groups"`) {
		t.Fatalf("nearby response: %d %s", w.Code, w.Body.String())
	}
	if rider.query.LimitPerMode != 2 || rider.query.Limit != 10 || rider.query.WheelchairAccessible == nil || !*rider.query.WheelchairAccessible {
		t.Fatalf("nearby query was not preserved: %#v", rider.query)
	}
	if w = request(h, "/v1/stops/nearby?latitude=91&longitude=-122.67"); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid bounds: %d", w.Code)
	}
	first := request(h, "/v1/stops/fixture:stop:1")
	if first.Code != http.StatusOK || first.Header().Get("X-Static-Feed-Version") != "fixture-v1" || first.Header().Get("ETag") == "" {
		t.Fatalf("stop: %d %#v", first.Code, first.Header())
	}
	r := httptest.NewRequest(http.MethodGet, "/v1/stops/fixture:stop:1", nil)
	r.Header.Set("If-None-Match", first.Header().Get("ETag"))
	r.RemoteAddr = "198.51.100.7:123"
	conditional := httptest.NewRecorder()
	h.ServeHTTP(conditional, r)
	if conditional.Code != http.StatusNotModified {
		t.Fatalf("stop etag: %d", conditional.Code)
	}
	if w = request(h, "/v1/routes/fixture:route:20"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"directions"`) {
		t.Fatalf("route: %d %s", w.Code, w.Body.String())
	}
	if w = request(h, "/v1/routes/fixture:route:20/shape"); w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/geo+json; charset=utf-8" {
		t.Fatalf("shape: %d %#v", w.Code, w.Header())
	}
	if w = request(h, "/v1/stops/fixture:stop:1/arrivals"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"status":"scheduled"`) {
		t.Fatalf("arrivals: %d %s", w.Code, w.Body.String())
	}
	if w = request(h, "/v1/stops/fixture:stop:1/schedule?serviceDate=2026-07-22"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"serviceDaySeconds":90930`) {
		t.Fatalf("schedule after midnight: %d %s", w.Code, w.Body.String())
	}
	if w = request(h, "/v1/alerts?effect=other"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"fixture:alert:one"`) {
		t.Fatalf("alerts: %d %s", w.Code, w.Body.String())
	}
}
