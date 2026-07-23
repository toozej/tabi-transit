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
func FuzzParseVehiclePositions(f *testing.F) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	f.Add(vehicleFeed(uint64(now.Unix()), false, 45.5))
	f.Add([]byte("bad"))
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = ParseVehiclePositions(raw, now, time.Minute, time.Minute) })
}
