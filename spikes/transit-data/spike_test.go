package transitdataspike

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gtfs "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

func ptr[T any](v T) *T { return &v }

func TestParseStaticGTFSAndAfterMidnightTimes(t *testing.T) {
	root := filepath.Join("fixtures", "gtfs")
	files := map[string]*os.File{}
	for _, name := range []string{"stops.txt", "routes.txt", "trips.txt", "stop_times.txt"} {
		f, err := os.Open(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		files[name] = f
	}
	readers := map[string]io.Reader{}
	for name, f := range files {
		readers[name] = f
	}
	feed, err := ParseStaticGTFS(readers)
	if err != nil {
		t.Fatal(err)
	}
	if len(feed.Stops) != 2 || len(feed.Routes) != 2 || len(feed.Trips) != 1 || len(feed.StopTimes) != 2 {
		t.Fatalf("unexpected feed counts: %#v", feed)
	}
	seconds, err := ServiceDaySeconds("25:15:30")
	if err != nil || seconds != 90930 {
		t.Fatalf("after-midnight conversion = %d, %v", seconds, err)
	}
	if _, err := ServiceDaySeconds("24:61:00"); err == nil {
		t.Fatal("expected invalid minute rejection")
	}
}

func TestParseGTFSRealtimeProtobuf(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("fixtures", "gtfsrt", "vehicle_positions.pb.base64"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(fixture)))
	if err != nil {
		t.Fatal(err)
	}
	vehicles, err := ParseVehiclePositions(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(vehicles) != 1 || vehicles[0].ID != "2901" || vehicles[0].Timestamp != 1_753_200_001 {
		t.Fatalf("unexpected decoded vehicles: %#v", vehicles)
	}
	invalid := &gtfs.FeedMessage{}
	if err := proto.Unmarshal(encoded, invalid); err != nil {
		t.Fatal(err)
	}
	invalid.Entity[0].Vehicle.Position.Latitude = ptr(float32(100))
	bad, err := proto.Marshal(invalid)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseVehiclePositions(bad); err == nil {
		t.Fatal("expected impossible-coordinate rejection")
	}
	if _, err := ParseVehiclePositions(bytes.Repeat([]byte{0xff}, 8)); err == nil {
		t.Fatal("expected malformed protobuf rejection")
	}
}
