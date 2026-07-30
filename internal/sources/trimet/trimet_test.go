package trimet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestLoadConfig(t *testing.T) {
	t.Parallel()
	secret := "never-log-this-AppID"
	config, err := LoadConfig(func(key string) string {
		switch key {
		case "TRIMET_ENABLED":
			return "true"
		case "TRIMET_BASE_URL":
			return "https://developer.trimet.org"
		case "TRIMET_APP_ID_FILE":
			return "/run/secrets/trimet"
		case "TRIMET_TIMEOUT":
			return "7s"
		}
		return ""
	}, func(path string) ([]byte, error) {
		if path != "/run/secrets/trimet" {
			t.Fatalf("unexpected secret path: %s", path)
		}
		return []byte(secret + "\n"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.AppID != secret || config.Timeout != 7*time.Second {
		t.Fatalf("unexpected config: %#v", config)
	}
	if _, err := LoadConfig(func(key string) string {
		if key == "TRIMET_ENABLED" {
			return "true"
		}
		if key == "TRIMET_BASE_URL" {
			return "https://not-trimet.invalid"
		}
		return ""
	}, func(string) ([]byte, error) { return nil, os.ErrNotExist }); err == nil {
		t.Fatal("expected invalid enabled config")
	}
	if _, err := LoadConfig(func(key string) string {
		if key == "TRIMET_PLANNER_ENABLED" {
			return "true"
		}
		return ""
	}, func(string) ([]byte, error) { return nil, nil }); err == nil {
		t.Fatal("expected planner disabled config error")
	}
}

func TestLoadConfigAcceptsRepositoryPrefixedEnvironment(t *testing.T) {
	t.Parallel()
	config, err := LoadConfig(func(key string) string {
		switch key {
		case "TABI_TRIMET_ENABLED":
			return "true"
		case "TABI_TRIMET_BASE_URL":
			return "https://developer.trimet.org"
		case "TABI_TRIMET_APP_ID":
			return "fixture-only"
		}
		return ""
	}, func(string) ([]byte, error) { return nil, os.ErrNotExist })
	if err != nil {
		t.Fatal(err)
	}
	if !config.Enabled || config.AppID != "fixture-only" || config.BaseURL != "https://developer.trimet.org" {
		t.Fatalf("unexpected prefixed config: %#v", config)
	}
}

func TestDisabledDoesNotCallProvider(t *testing.T) {
	t.Parallel()
	client, err := NewClient(Config{Timeout: time.Second}, &http.Client{}, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = client.Arrivals(context.Background(), ArrivalsRequest{StopID: "8334", Minutes: 10})
	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("got %v", err)
	}
	if IsSourceUnavailable(err) {
		t.Fatal("disabled source must not be treated as transient provider outage")
	}
}

func TestArrivalsMapsFixtureAndKeepsCredentialOutOfErrors(t *testing.T) {
	t.Parallel()
	fixture, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "upstream", "trimet", "arrivals.json"))
	if err != nil {
		t.Fatal(err)
	}
	appID := "super-secret-app-id"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/ws/v2/arrivals" {
			t.Errorf("path: %s", request.URL.Path)
		}
		if got := request.URL.Query().Get("appID"); got != appID {
			t.Errorf("AppID not injected")
		}
		if request.URL.Query().Get("locIDs") != "8334" {
			t.Errorf("locIDs missing")
		}
		if request.URL.Query().Get("json") != "true" {
			t.Errorf("json response was not requested")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	client, err := NewClient(Config{Enabled: true, AppID: appID, BaseURL: server.URL, AllowedHosts: []string{u.Hostname()}, Timeout: time.Second}, server.Client(), fixedClock{now: time.Date(2026, 7, 22, 16, 30, 2, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	arrivals, freshness, err := client.Arrivals(context.Background(), ArrivalsRequest{StopID: "8334", Minutes: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(arrivals) != 1 || arrivals[0].VehicleID != "2901" || arrivals[0].EstimatedAt == nil || !arrivals[0].Streetcar {
		t.Fatalf("unexpected arrivals: %#v", arrivals)
	}
	if freshness.Source != SourceID || !freshness.IsRealtime || freshness.FetchedAt.IsZero() {
		t.Fatalf("unexpected freshness: %#v", freshness)
	}
	if strings.Contains(errString(err), appID) || strings.Contains(fmt.Sprint(errors.Unwrap(err)), appID) {
		t.Fatal("AppID leaked through error")
	}
}

func TestArrivalsAcceptsDocumentedEpochMilliseconds(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"resultSet":{"queryTime":1784737802000,"arrival":[{"locid":"8334","route":"20","scheduled":1784738100000,"estimated":1784738160000,"status":"estimated"}],"newField":"provider-compatible"}}`))
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	client, err := NewClient(Config{Enabled: true, AppID: "fixture-only", BaseURL: server.URL, AllowedHosts: []string{u.Hostname()}, Timeout: time.Second}, server.Client(), fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	arrivals, freshness, err := client.Arrivals(context.Background(), ArrivalsRequest{StopID: "8334", Minutes: 60})
	if err != nil {
		t.Fatal(err)
	}
	if len(arrivals) != 1 || arrivals[0].EstimatedAt == nil || arrivals[0].EstimatedAt.UnixMilli() != 1784738160000 {
		t.Fatalf("unexpected mapped arrivals: %#v", arrivals)
	}
	if freshness.SourceUpdatedAt == nil || freshness.SourceUpdatedAt.UnixMilli() != 1784737802000 {
		t.Fatalf("unexpected source freshness: %#v", freshness)
	}
	if _, _, err := client.Arrivals(context.Background(), ArrivalsRequest{StopID: "8334", Minutes: 61}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected minutes bound error, got %v", err)
	}
}

func TestArrivalsNormalizesNumericProviderIDs(t *testing.T) {
	t.Parallel()
	input := arrivalsResponse{}
	input.ResultSet.Arrival = []arrivalDTO{{StopID: providerID(`8334`), RouteID: providerID(`20`)}}
	arrivals := mapArrivals(input)
	if len(arrivals) != 1 || arrivals[0].StopID != "8334" || arrivals[0].RouteID != "20" {
		t.Fatalf("unexpected numeric ID normalization: %#v", arrivals)
	}
}

func TestClassifiesProviderFailuresWithoutBodyLeakage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		handler     http.HandlerFunc
		want        ErrorKind
		unavailable bool
	}{
		{"non-2xx", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("credential=should-not-leak"))
		}, ErrorUnavailable, true},
		{"malformed", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{not-json")) }, ErrorMalformed, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewTLSServer(tc.handler)
			defer server.Close()
			u, _ := url.Parse(server.URL)
			appID := "very-secret"
			client, err := NewClient(Config{Enabled: true, AppID: appID, BaseURL: server.URL, AllowedHosts: []string{u.Hostname()}, Timeout: time.Second}, server.Client(), fixedClock{})
			if err != nil {
				t.Fatal(err)
			}
			_, _, err = client.Arrivals(context.Background(), ArrivalsRequest{StopID: "8334"})
			var sourceErr *Error
			if !errors.As(err, &sourceErr) || sourceErr.Kind != tc.want {
				t.Fatalf("got %v", err)
			}
			if IsSourceUnavailable(err) != tc.unavailable {
				t.Fatalf("unavailable mapping: %v", err)
			}
			if strings.Contains(err.Error(), appID) || strings.Contains(fmt.Sprint(errors.Unwrap(err)), appID) || strings.Contains(err.Error(), "credential") {
				t.Fatalf("secret/provider body leaked: %v", err)
			}
		})
	}
}

func TestTimeoutAndInvalidRequests(t *testing.T) {
	t.Parallel()
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { time.Sleep(100 * time.Millisecond) }))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	client, err := NewClient(Config{Enabled: true, AppID: "secret", BaseURL: server.URL, AllowedHosts: []string{u.Hostname()}, Timeout: time.Second}, server.Client(), fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, _, err = client.Arrivals(ctx, ArrivalsRequest{StopID: "8334"})
	var sourceErr *Error
	if !errors.As(err, &sourceErr) || sourceErr.Kind != ErrorTimeout || !IsSourceUnavailable(err) {
		t.Fatalf("got %v", err)
	}
	_, _, err = client.Arrivals(context.Background(), ArrivalsRequest{StopID: "bad\nstop"})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("got %v", err)
	}
}

func TestPlannerFeatureGate(t *testing.T) {
	t.Parallel()
	client, err := NewClient(Config{Timeout: time.Second}, nil, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = client.Plan(context.Background(), PlanRequest{Origin: "8334", Destination: "1000"})
	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("got %v", err)
	}
}

func TestPlannerMapsSanitizedFixtureAndConstraints(t *testing.T) {
	t.Parallel()
	fixture, err := os.ReadFile(filepath.Join("..", "..", "..", "tests", "fixtures", "upstream", "trimet", "trip_planner.json"))
	if err != nil {
		t.Fatal(err)
	}
	maxTransfers, maxWalk := 1, 1200
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/ws/v2/tripplanner" {
			t.Errorf("path: %s", request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("fromPlace") != "tabi:stop:8334" || query.Get("toPlace") != "tabi:stop:1000" {
			t.Errorf("unexpected places: %q %q", query.Get("fromPlace"), query.Get("toPlace"))
		}
		if query.Get("modes") != "bus,walk" || query.Get("maxTransfers") != "1" || query.Get("maxWalkMeters") != "1200" || query.Get("accessible") != "true" || query.Get("arriveBy") != "true" {
			t.Errorf("constraints not mapped: %v", query)
		}
		_, _ = w.Write(fixture)
	}))
	defer server.Close()
	u, _ := url.Parse(server.URL)
	client, err := NewClient(Config{Enabled: true, PlannerEnabled: true, AppID: "fixture-only", BaseURL: server.URL, AllowedHosts: []string{u.Hostname()}, Timeout: time.Second}, server.Client(), fixedClock{now: time.Date(2026, 7, 23, 16, 32, 1, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	plan, freshness, err := client.Plan(context.Background(), PlanRequest{
		Origin: "tabi:stop:8334", Destination: "tabi:stop:1000", ArriveBy: true,
		Preferences: PlanPreferences{Modes: []Mode{ModeBus, ModeWalk}, MaxTransfers: &maxTransfers, MaxWalkMeters: &maxWalk, RequireAccessibility: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.ID != "fixture-plan-20" || len(plan.Itineraries) != 1 || len(plan.Itineraries[0].Legs) != 2 || plan.Itineraries[0].Legs[1].Mode != ModeBus || plan.Itineraries[0].Legs[1].DistanceMeters == nil {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if freshness.Source != SourceID || !freshness.IsRealtime {
		t.Fatalf("unexpected freshness: %#v", freshness)
	}
}

func TestPlannerRejectsUnsafeConstraints(t *testing.T) {
	t.Parallel()
	client, err := NewClient(Config{Enabled: true, PlannerEnabled: true, AppID: "fixture-only", BaseURL: "https://developer.trimet.org", AllowedHosts: []string{"developer.trimet.org"}, Timeout: time.Second}, nil, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	negativeWalk := -1
	_, _, err = client.Plan(context.Background(), PlanRequest{Origin: "a", Destination: "b", Preferences: PlanPreferences{Modes: []Mode{Mode("scooter")}, MaxWalkMeters: &negativeWalk}})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("got %v", err)
	}
}

func TestConfigRejectsBaseURLCredentialsOrQuery(t *testing.T) {
	t.Parallel()
	for _, baseURL := range []string{"https://user:pass@developer.trimet.org", "https://developer.trimet.org?unexpected=true"} {
		if err := (Config{Enabled: true, AppID: "fixture-only", BaseURL: baseURL, AllowedHosts: []string{"developer.trimet.org"}, Timeout: time.Second}).Validate(); err == nil {
			t.Fatalf("expected invalid base URL %q", baseURL)
		}
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
