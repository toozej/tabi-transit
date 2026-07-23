package poller

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	gtfs "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/toozej/tabi-transit/internal/persistence"
	"google.golang.org/protobuf/proto"
)

type memoryStore struct {
	mu          sync.Mutex
	snapshots   []persistence.VehicleSnapshot
	failures    []string
	tripUpdates []persistence.TripUpdateSnapshot
}

func (s *memoryStore) ReplaceTripUpdateSnapshot(_ context.Context, snapshot persistence.TripUpdateSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tripUpdates = append(s.tripUpdates, snapshot)
	return nil
}

func (s *memoryStore) ReplaceVehicleSnapshot(_ context.Context, snapshot persistence.VehicleSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots = append(s.snapshots, snapshot)
	return nil
}
func (s *memoryStore) RecordSourceFailure(_ context.Context, _ string, code string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures = append(s.failures, code)
	return nil
}
func ps(v string) *string   { return &v }
func pu(v uint64) *uint64   { return &v }
func pf(v float32) *float32 { return &v }
func validPayload(now time.Time) []byte {
	m := &gtfs.FeedMessage{Header: &gtfs.FeedHeader{GtfsRealtimeVersion: ps("2.0"), Timestamp: pu(uint64(now.Unix()))}, Entity: []*gtfs.FeedEntity{{Id: ps("e"), Vehicle: &gtfs.VehiclePosition{Vehicle: &gtfs.VehicleDescriptor{Id: ps("2901")}, Position: &gtfs.Position{Latitude: pf(45.5), Longitude: pf(-122.6)}, Trip: &gtfs.TripDescriptor{RouteId: ps("20"), TripId: ps("T")}, Timestamp: pu(uint64(now.Unix()))}}}}
	raw, _ := proto.Marshal(m)
	return raw
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
func service(transport roundTrip, store *memoryStore, now time.Time) Service {
	return Service{Config: Config{SourceID: "trimet", Endpoint: "https://fixtures.invalid/vehicle-positions", AllowedHosts: []string{"fixtures.invalid"}, Timeout: time.Second, Interval: time.Millisecond, StaleAfter: time.Minute, MaxBytes: 1024}, HTTPClient: &http.Client{Transport: transport}, Store: store, Clock: func() time.Time { return now }}
}
func TestValidSnapshotUsesSourceQualifiedIDs(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	store := &memoryStore{}
	transport := roundTrip(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(validPayload(now))))}, nil
	})
	if err := service(transport, store, now).Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(store.snapshots) != 1 || store.snapshots[0].Vehicles[0].ID != "trimet:vehicle:2901" || *store.snapshots[0].Vehicles[0].RouteID != "trimet:route:20" {
		t.Fatalf("bad snapshot: %#v", store.snapshots)
	}
}
func TestBadOrEmptyDoesNotReplaceSnapshot(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	cases := []struct{ name, body, code string }{{"malformed", "bad", "validation_failed"}, {"empty", "", "validation_failed"}}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := &memoryStore{snapshots: []persistence.VehicleSnapshot{{SourceID: "prior"}}}
			transport := roundTrip(func(_ *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			err := service(transport, store, now).Run(context.Background())
			if err == nil || len(store.snapshots) != 1 || len(store.failures) != 1 || store.failures[0] != tc.code {
				t.Fatalf("err=%v snapshots=%d failures=%v", err, len(store.snapshots), store.failures)
			}
		})
	}
}
func TestConfigDisabledAndAllowlisted(t *testing.T) {
	t.Parallel()
	if _, err := LoadConfig(func(string) string { return "" }, func(string) ([]byte, error) { return nil, nil }); !errors.Is(err, ErrDisabled) {
		t.Fatalf("got %v", err)
	}
	if err := (Config{SourceID: "trimet", Endpoint: "https://example.invalid/feed", AllowedHosts: []string{"wrong.invalid"}, Timeout: time.Second, Interval: time.Second, StaleAfter: time.Minute, MaxBytes: 1}).Validate(); err == nil || !strings.Contains(err.Error(), "allowlisted") {
		t.Fatal(err)
	}
}

func TestTripUpdateOnlyConfigIsExplicitAndAllowlisted(t *testing.T) {
	t.Parallel()
	c, err := LoadConfig(func(key string) string {
		if key == "GTFSRT_TRIP_UPDATES_ENDPOINT" {
			return "https://fixtures.invalid/trip-updates"
		}
		if key == "GTFSRT_ALLOWED_HOSTS" {
			return "fixtures.invalid"
		}
		return ""
	}, func(string) ([]byte, error) { return nil, nil })
	if err != nil || c.Endpoint != "" || c.TripUpdatesEndpoint == "" {
		t.Fatalf("config=%#v err=%v", c, err)
	}
	if err := c.Validate(); !errors.Is(err, ErrDisabled) {
		t.Fatalf("vehicle validation %v", err)
	}
	if err := c.ValidateTripUpdates(); err != nil {
		t.Fatal(err)
	}
}

func TestVehicleOnlyConfigRemainsValid(t *testing.T) {
	t.Parallel()
	c, err := LoadConfig(func(key string) string {
		if key == "GTFSRT_VEHICLE_ENDPOINT" {
			return "https://fixtures.invalid/vehicles"
		}
		if key == "GTFSRT_ALLOWED_HOSTS" {
			return "fixtures.invalid"
		}
		return ""
	}, func(string) ([]byte, error) { return nil, nil })
	if err != nil || c.TripUpdatesEndpoint != "" {
		t.Fatalf("config=%#v err=%v", c, err)
	}
}
func TestCancellationAndTimeoutPreserveSnapshot(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	store := &memoryStore{}
	s := service(roundTrip(func(*http.Request) (*http.Response, error) { return nil, context.Canceled }), store, now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Run(ctx); err == nil || len(store.snapshots) != 0 || len(store.failures) != 1 {
		t.Fatalf("%v %#v", err, store)
	}
}

func TestTripUpdatesAreNormalizedAndBadPayloadPreservesCurrentState(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	valid := &gtfs.FeedMessage{
		Header: &gtfs.FeedHeader{GtfsRealtimeVersion: ps("2.0"), Timestamp: pu(uint64(now.Unix()))},
		Entity: []*gtfs.FeedEntity{{
			Id: ps("tu"),
			TripUpdate: &gtfs.TripUpdate{
				Trip: &gtfs.TripDescriptor{TripId: ps("T"), RouteId: ps("20")},
				StopTimeUpdate: []*gtfs.TripUpdate_StopTimeUpdate{{
					StopSequence: proto.Uint32(1), StopId: ps("S"), Arrival: &gtfs.TripUpdate_StopTimeEvent{Delay: proto.Int32(60)},
				}},
			},
		}},
	}
	raw, _ := proto.Marshal(valid)
	store := &memoryStore{}
	transport := roundTrip(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(raw)))}, nil
	})
	s := service(transport, store, now)
	s.Config.TripUpdatesEndpoint = "https://fixtures.invalid/trip-updates"
	if err := s.RunTripUpdates(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(store.tripUpdates) != 1 || store.tripUpdates[0].Updates[0].TripID != "trimet:trip:T" || store.tripUpdates[0].Updates[0].StopTimes[0].StopID != "trimet:stop:S" {
		t.Fatalf("updates=%#v", store.tripUpdates)
	}
	bad := service(roundTrip(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("bad"))}, nil
	}), store, now)
	bad.Config.TripUpdatesEndpoint = "https://fixtures.invalid/trip-updates"
	if err := bad.RunTripUpdates(context.Background()); err == nil || len(store.tripUpdates) != 1 || len(store.failures) != 1 {
		t.Fatalf("err=%v updates=%d failures=%v", err, len(store.tripUpdates), store.failures)
	}
}
