package gtfsrt

import (
	"errors"
	"testing"
	"time"

	gtfs "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

func str(v string) *string   { return &v }
func u64(v uint64) *uint64   { return &v }
func f32(v float32) *float32 { return &v }
func vehicleFeed(ts uint64, deleted bool, lat float32) []byte {
	message := &gtfs.FeedMessage{Header: &gtfs.FeedHeader{GtfsRealtimeVersion: str("2.0"), Timestamp: u64(ts)}, Entity: []*gtfs.FeedEntity{{Id: str("entity-1"), IsDeleted: &deleted, Vehicle: &gtfs.VehiclePosition{Vehicle: &gtfs.VehicleDescriptor{Id: str("2901")}, Position: &gtfs.Position{Latitude: f32(lat), Longitude: f32(-122.67)}, Trip: &gtfs.TripDescriptor{RouteId: str("20"), TripId: str("trip-1")}, Timestamp: u64(ts)}}}}
	raw, _ := proto.Marshal(message)
	return raw
}
func TestParseVehiclePositions(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	feed, err := ParseVehiclePositions(vehicleFeed(uint64(now.Unix()), false, 45.5), now, time.Minute, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Vehicles) != 1 || feed.Vehicles[0].SourceVehicleID != "2901" || feed.Vehicles[0].RouteID != "20" || len(feed.SHA256) != 64 {
		t.Fatalf("unexpected feed: %#v", feed)
	}
}
func TestParseRejectsUnsafeSnapshots(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		raw  []byte
		want error
	}{
		{"malformed", []byte("not protobuf"), ErrMalformed},
		{"deleted", vehicleFeed(uint64(now.Unix()), true, 45.5), ErrEmpty},
		{"stale", vehicleFeed(uint64(now.Add(-2*time.Minute).Unix()), false, 45.5), ErrMalformed},
		{"coordinate", vehicleFeed(uint64(now.Unix()), false, 91), ErrMalformed},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseVehiclePositions(tc.raw, now, time.Minute, time.Minute)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v", err)
			}
		})
	}
}
func TestParseRejectsDifferentialAndDuplicateVehicleIDs(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	var differential gtfs.FeedMessage
	if err := proto.Unmarshal(vehicleFeed(uint64(now.Unix()), false, 45.5), &differential); err != nil {
		t.Fatal(err)
	}
	differential.Header.Incrementality = gtfs.FeedHeader_DIFFERENTIAL.Enum()
	raw, _ := proto.Marshal(&differential)
	if _, err := ParseVehiclePositions(raw, now, time.Minute, time.Minute); !errors.Is(err, ErrMalformed) {
		t.Fatalf("differential: %v", err)
	}
	differential.Header.Incrementality = gtfs.FeedHeader_FULL_DATASET.Enum()
	differential.Entity = append(differential.Entity, proto.Clone(differential.Entity[0]).(*gtfs.FeedEntity))
	raw, _ = proto.Marshal(&differential)
	if _, err := ParseVehiclePositions(raw, now, time.Minute, time.Minute); !errors.Is(err, ErrMalformed) {
		t.Fatalf("duplicate: %v", err)
	}
}
func TestParseTripUpdatesAndAlerts(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	message := &gtfs.FeedMessage{Header: &gtfs.FeedHeader{GtfsRealtimeVersion: str("2.0"), Timestamp: u64(uint64(now.Unix()))}, Entity: []*gtfs.FeedEntity{
		{Id: str("trip-entity"), TripUpdate: &gtfs.TripUpdate{
			Trip:           &gtfs.TripDescriptor{TripId: str("trip-1"), RouteId: str("20"), StartDate: str("20260722")},
			StopTimeUpdate: []*gtfs.TripUpdate_StopTimeUpdate{{StopSequence: proto.Uint32(1), StopId: str("stop-1"), Arrival: &gtfs.TripUpdate_StopTimeEvent{Delay: proto.Int32(90)}}},
		}},
		{Id: str("alert-entity"), Alert: &gtfs.Alert{
			HeaderText:   &gtfs.TranslatedString{Translation: []*gtfs.TranslatedString_Translation{{Text: str("Fixture alert")}}},
			ActivePeriod: []*gtfs.TimeRange{{Start: u64(uint64(now.Add(-time.Minute).Unix())), End: u64(uint64(now.Add(time.Hour).Unix()))}},
		}},
	}}
	raw, err := proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	updates, err := ParseTripUpdates(raw, now, time.Minute, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(updates) != 1 || updates[0].TripID != "trip-1" || updates[0].StopTimes[0].ArrivalDelaySeconds == nil {
		t.Fatalf("updates: %#v", updates)
	}
	alerts, err := ParseAlerts(raw, now, time.Minute, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 || alerts[0].Header != "Fixture alert" || alerts[0].ActiveUntil == nil {
		t.Fatalf("alerts: %#v", alerts)
	}
}
func TestTripUpdatesAndAlertsRejectMalformedOrStale(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	stale := vehicleFeed(uint64(now.Add(-2*time.Minute).Unix()), false, 45.5)
	for name, parse := range map[string]func([]byte, time.Time, time.Duration, time.Duration) error{
		"trip": func(raw []byte, n time.Time, age, skew time.Duration) error {
			_, err := ParseTripUpdates(raw, n, age, skew)
			return err
		},
		"alert": func(raw []byte, n time.Time, age, skew time.Duration) error {
			_, err := ParseAlerts(raw, n, age, skew)
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := parse(stale, now, time.Minute, time.Minute); !errors.Is(err, ErrMalformed) {
				t.Fatalf("stale: %v", err)
			}
			if err := parse([]byte("bad"), now, time.Minute, time.Minute); !errors.Is(err, ErrMalformed) {
				t.Fatalf("malformed: %v", err)
			}
		})
	}
}
func FuzzParseVehiclePositions(f *testing.F) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	f.Add(vehicleFeed(uint64(now.Unix()), false, 45.5))
	f.Add([]byte("bad"))
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = ParseVehiclePositions(raw, now, time.Minute, time.Minute) })
}
