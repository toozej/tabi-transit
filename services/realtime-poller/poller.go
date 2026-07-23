// Package poller fetches approved GTFS-Realtime vehicle-position feeds.
package poller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/toozej/tabi-transit/internal/persistence"
	"github.com/toozej/tabi-transit/internal/sources/gtfsrt"
)

var ErrDisabled = errors.New("GTFS-Realtime vehicle endpoint is not configured")

type Config struct {
	SourceID, Endpoint, EndpointFile, TripUpdatesEndpoint, TripUpdatesEndpointFile string
	AllowedHosts                                                                   []string
	Timeout, Interval, StaleAfter                                                  time.Duration
	MaxBytes                                                                       int64
}

func LoadConfig(getenv func(string) string, readFile func(string) ([]byte, error)) (Config, error) {
	c := Config{SourceID: strings.TrimSpace(getenv("GTFSRT_SOURCE_ID")), Endpoint: strings.TrimSpace(getenv("GTFSRT_VEHICLE_ENDPOINT")), EndpointFile: strings.TrimSpace(getenv("GTFSRT_VEHICLE_ENDPOINT_FILE")), TripUpdatesEndpoint: strings.TrimSpace(getenv("GTFSRT_TRIP_UPDATES_ENDPOINT")), TripUpdatesEndpointFile: strings.TrimSpace(getenv("GTFSRT_TRIP_UPDATES_ENDPOINT_FILE")), AllowedHosts: splitCSV(getenv("GTFSRT_ALLOWED_HOSTS")), Timeout: 10 * time.Second, Interval: 30 * time.Second, StaleAfter: 2 * time.Minute, MaxBytes: 5 << 20}
	if c.SourceID == "" {
		c.SourceID = "trimet"
	}
	for _, setting := range []struct {
		name, value string
		target      *time.Duration
	}{{"GTFSRT_HTTP_TIMEOUT", getenv("GTFSRT_HTTP_TIMEOUT"), &c.Timeout}, {"GTFSRT_INTERVAL", getenv("GTFSRT_INTERVAL"), &c.Interval}, {"GTFSRT_STALE_AFTER", getenv("GTFSRT_STALE_AFTER"), &c.StaleAfter}} {
		if strings.TrimSpace(setting.value) != "" {
			d, err := time.ParseDuration(setting.value)
			if err != nil {
				return Config{}, fmt.Errorf("%s: %w", setting.name, err)
			}
			*setting.target = d
		}
	}
	if c.Endpoint == "" && c.EndpointFile != "" {
		raw, err := readFile(c.EndpointFile)
		if err != nil {
			return Config{}, fmt.Errorf("GTFSRT_VEHICLE_ENDPOINT_FILE: %w", err)
		}
		c.Endpoint = strings.TrimSpace(string(raw))
	}
	if c.TripUpdatesEndpoint == "" && c.TripUpdatesEndpointFile != "" {
		raw, err := readFile(c.TripUpdatesEndpointFile)
		if err != nil {
			return Config{}, fmt.Errorf("GTFSRT_TRIP_UPDATES_ENDPOINT_FILE: %w", err)
		}
		c.TripUpdatesEndpoint = strings.TrimSpace(string(raw))
	}
	return c, c.ValidateAny()
}
func (c Config) Validate() error {
	if c.Endpoint == "" {
		return ErrDisabled
	}
	return c.validateEndpoint(c.Endpoint, "GTFSRT_VEHICLE_ENDPOINT")
}
func (c Config) ValidateTripUpdates() error {
	if c.TripUpdatesEndpoint == "" {
		return ErrDisabled
	}
	return c.validateEndpoint(c.TripUpdatesEndpoint, "GTFSRT_TRIP_UPDATES_ENDPOINT")
}
func (c Config) ValidateAny() error {
	if c.Endpoint == "" && c.TripUpdatesEndpoint == "" {
		return ErrDisabled
	}
	if c.Endpoint != "" {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	if c.TripUpdatesEndpoint != "" {
		return c.ValidateTripUpdates()
	}
	return nil
}
func (c Config) validateEndpoint(endpoint, name string) error {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" {
		return fmt.Errorf("%s must be an HTTPS URL", name)
	}
	if c.SourceID == "" || strings.ContainsAny(c.SourceID, "\r\n:\t") {
		return errors.New("GTFSRT_SOURCE_ID is invalid")
	}
	allowed := false
	for _, host := range c.AllowedHosts {
		if strings.EqualFold(host, u.Hostname()) {
			allowed = true
		}
	}
	if !allowed {
		return fmt.Errorf("%s host is not allowlisted", name)
	}
	if c.Timeout <= 0 || c.Timeout > 60*time.Second || c.Interval <= 0 || c.StaleAfter <= 0 || c.MaxBytes < 1 || c.MaxBytes > 20<<20 {
		return errors.New("invalid GTFS-Realtime poller bounds")
	}
	return nil
}
func splitCSV(raw string) []string {
	var out []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

type Store interface {
	ReplaceVehicleSnapshot(context.Context, persistence.VehicleSnapshot) error
	ReplaceTripUpdateSnapshot(context.Context, persistence.TripUpdateSnapshot) error
	RecordSourceFailure(context.Context, string, string, time.Time) error
}
type Service struct {
	Config     Config
	HTTPClient *http.Client
	Store      Store
	Clock      func() time.Time
}

func (s Service) Run(ctx context.Context) error {
	if err := s.Config.Validate(); err != nil {
		return err
	}
	if s.Store == nil {
		return errors.New("realtime poller store is not configured")
	}
	now := time.Now
	if s.Clock != nil {
		now = s.Clock
	}
	fetched := now().UTC()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Config.Endpoint, nil)
	if err != nil {
		return err
	}
	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: s.Config.Timeout}
	}
	response, err := client.Do(request)
	if err != nil {
		return s.failure(ctx, "fetch_failed", fetched, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return s.failure(ctx, "fetch_status", fetched, fmt.Errorf("unexpected status %d", response.StatusCode))
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, s.Config.MaxBytes+1))
	if err != nil || int64(len(raw)) > s.Config.MaxBytes {
		return s.failure(ctx, "payload_invalid", fetched, errors.New("GTFS-Realtime payload exceeds bound or cannot be read"))
	}
	feed, err := gtfsrt.ParseVehiclePositions(raw, fetched, s.Config.StaleAfter, 5*time.Minute)
	if err != nil {
		return s.failure(ctx, code(err), fetched, err)
	}
	vehicles := make([]persistence.Vehicle, 0, len(feed.Vehicles))
	for _, vehicle := range feed.Vehicles {
		vehicles = append(vehicles, persistence.Vehicle{ID: public(s.Config.SourceID, "vehicle", vehicle.SourceVehicleID), SourceID: s.Config.SourceID, SourceVehicleID: vehicle.SourceVehicleID, RouteID: publicPtr(s.Config.SourceID, "route", vehicle.RouteID), TripID: publicPtr(s.Config.SourceID, "trip", vehicle.TripID), Mode: "unknown", Coordinate: persistence.Coordinate{Longitude: vehicle.Longitude, Latitude: vehicle.Latitude}, SourceUpdatedAt: &feed.SourceUpdatedAt, EntityUpdatedAt: vehicle.EntityUpdatedAt, FetchedAt: fetched, ProcessedAt: now().UTC(), Freshness: freshness(fetched.Sub(feed.SourceUpdatedAt), s.Config.StaleAfter)})
	}
	if err := s.Store.ReplaceVehicleSnapshot(ctx, persistence.VehicleSnapshot{SourceID: s.Config.SourceID, SourceUpdatedAt: &feed.SourceUpdatedAt, FetchedAt: fetched, ProcessedAt: now().UTC(), ContentSHA256: feed.SHA256, Vehicles: vehicles}); err != nil {
		return s.failure(ctx, "store_failed", fetched, err)
	}
	return nil
}
func (s Service) RunLoop(ctx context.Context) error {
	for {
		if err := s.Run(ctx); err != nil && errors.Is(err, ErrDisabled) {
			return err
		}
		timer := time.NewTimer(s.Config.Interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// RunTripUpdates uses the same bounded, allowlisted transport policy as
// vehicle positions, but keeps the projection separate so a bad trip-update
// payload cannot disturb vehicle state. It is only invoked by explicit binary
// wiring; no endpoint is enabled by default.
func (s Service) RunTripUpdates(ctx context.Context) error {
	if err := s.Config.ValidateTripUpdates(); err != nil {
		return err
	}
	if s.Store == nil {
		return errors.New("realtime poller store is not configured")
	}
	now := time.Now
	if s.Clock != nil {
		now = s.Clock
	}
	fetched := now().UTC()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Config.TripUpdatesEndpoint, nil)
	if err != nil {
		return err
	}
	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: s.Config.Timeout}
	}
	response, err := client.Do(request)
	if err != nil {
		return s.failure(ctx, "fetch_failed", fetched, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return s.failure(ctx, "fetch_status", fetched, fmt.Errorf("unexpected status %d", response.StatusCode))
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, s.Config.MaxBytes+1))
	if err != nil || int64(len(raw)) > s.Config.MaxBytes {
		return s.failure(ctx, "payload_invalid", fetched, errors.New("GTFS-Realtime payload exceeds bound or cannot be read"))
	}
	feed, err := gtfsrt.ParseTripUpdateFeed(raw, fetched, s.Config.StaleAfter, 5*time.Minute)
	if err != nil {
		return s.failure(ctx, code(err), fetched, err)
	}
	normalized := make([]persistence.TripUpdate, 0, len(feed.Updates))
	for _, update := range feed.Updates {
		item := persistence.TripUpdate{EntityID: update.EntityID, TripID: public(s.Config.SourceID, "trip", update.TripID), RouteID: valueOrEmpty(publicPtr(s.Config.SourceID, "route", update.RouteID)), StartDate: update.StartDate, ScheduleRelationship: update.ScheduleRelationship, UpdatedAt: update.UpdatedAt}
		for _, stop := range update.StopTimes {
			item.StopTimes = append(item.StopTimes, persistence.TripUpdateStopTime{StopSequence: int(stop.StopSequence), StopID: valueOrEmpty(publicPtr(s.Config.SourceID, "stop", stop.StopID)), ArrivalDelaySeconds: stop.ArrivalDelaySeconds, ArrivalTime: stop.ArrivalTime, DepartureDelaySeconds: stop.DepartureDelaySeconds, DepartureTime: stop.DepartureTime, ScheduleRelationship: stop.ScheduleRelationship})
		}
		normalized = append(normalized, item)
	}
	if err := s.Store.ReplaceTripUpdateSnapshot(ctx, persistence.TripUpdateSnapshot{SourceID: s.Config.SourceID, SourceUpdatedAt: &feed.SourceUpdatedAt, FetchedAt: fetched, ProcessedAt: now().UTC(), ContentSHA256: feed.SHA256, Updates: normalized}); err != nil {
		return s.failure(ctx, "store_failed", fetched, err)
	}
	return nil
}
func (s Service) RunTripUpdatesLoop(ctx context.Context) error {
	for {
		if err := s.RunTripUpdates(ctx); err != nil && errors.Is(err, ErrDisabled) {
			return err
		}
		timer := time.NewTimer(s.Config.Interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
func (s Service) failure(ctx context.Context, safeCode string, at time.Time, cause error) error {
	_ = s.Store.RecordSourceFailure(ctx, s.Config.SourceID, safeCode, at)
	return fmt.Errorf("GTFS-Realtime poll %s: %w", safeCode, cause)
}
func public(source, kind, raw string) string { return source + ":" + kind + ":" + raw }
func publicPtr(source, kind, raw string) *string {
	if raw == "" {
		return nil
	}
	value := public(source, kind, raw)
	return &value
}
func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func freshness(age, stale time.Duration) persistence.FreshnessStatus {
	if age < 0 {
		return persistence.FreshnessUnknown
	}
	if age > stale {
		return persistence.FreshnessStale
	}
	if age > stale/2 {
		return persistence.FreshnessAging
	}
	return persistence.FreshnessFresh
}
func code(err error) string {
	if errors.Is(err, gtfsrt.ErrEmpty) {
		return "empty_snapshot"
	}
	return "validation_failed"
}

// DefaultConfig uses process environment only at the binary boundary.
func DefaultConfig() (Config, error) { return LoadConfig(os.Getenv, os.ReadFile) }
