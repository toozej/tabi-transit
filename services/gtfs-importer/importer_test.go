package importer

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"github.com/toozej/tabi-transit/internal/sources/gtfs"
	"io"
	"net/http"
	"testing"
	"time"
)

type memoryStore struct {
	imports  int
	failures int
	active   string
}

func (m *memoryStore) Import(_ context.Context, _, _, digest string, _ time.Time, _ gtfs.Feed) (bool, error) {
	if m.active == digest {
		return true, nil
	}
	m.imports++
	m.active = digest
	return false, nil
}
func (m *memoryStore) RecordFailure(context.Context, string, string, time.Time) error {
	m.failures++
	return nil
}
func fixture(t *testing.T) []byte {
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for n, v := range map[string]string{"stops.txt": "stop_id,stop_name,stop_lon,stop_lat\ns,Stop,-122,45\n", "routes.txt": "route_id,route_type\nr,3\n", "trips.txt": "route_id,service_id,trip_id\nr,d,t\n", "calendar_dates.txt": "service_id,date,exception_type\nd,20260101,1\n", "stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nt,25:00:00,25:01:00,s,1\n"} {
		f, _ := w.Create(n)
		_, _ = f.Write([]byte(v))
	}
	_ = w.Close()
	return b.Bytes()
}

type fixtureTransport struct{ body []byte }

func (f fixtureTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(f.body)), Header: make(http.Header)}, nil
}
func TestRunIdempotentAndFailurePreservesActive(t *testing.T) {
	body := fixture(t)
	store := &memoryStore{}
	s := Service{Config: Config{SourceID: "fixture", Endpoint: "https://fixture.invalid/gtfs.zip", AllowedHosts: []string{"fixture.invalid"}, Timeout: time.Second, Policy: DefaultPolicy()}, HTTPClient: &http.Client{Transport: fixtureTransport{body}}, Store: store}
	if e := s.Run(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e := s.Run(context.Background()); e != nil {
		t.Fatal(e)
	}
	if store.imports != 1 {
		t.Fatalf("imports %d", store.imports)
	}
	s.Config.ArchiveSHA256 = "nope"
	if e := s.Run(context.Background()); e == nil || store.active == "" || store.failures != 1 {
		t.Fatalf("failed import changed active or wasn't reported: %+v", store)
	}
}
func TestConfigIsDisabledWithoutEndpoint(t *testing.T) {
	e := Config{SourceID: "fixture"}.Validate()
	if !errors.Is(e, ErrDisabled) {
		t.Fatalf("%v", e)
	}
}

func TestRouteTypeFiveNormalizesToStreetcar(t *testing.T) {
	if got := mode("5"); got != "streetcar" {
		t.Fatalf("GTFS route_type=5 mode = %q, want streetcar", got)
	}
}
