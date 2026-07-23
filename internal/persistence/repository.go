// Package persistence defines storage boundaries shared by Tabi's binaries.
// It deliberately exposes source-qualified public identifiers, never database
// surrogate keys, to keep provider mappings behind importer and poller code.
package persistence

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalidPublicID = errors.New("invalid source-qualified public ID")

type FreshnessStatus string

const (
	FreshnessFresh   FreshnessStatus = "fresh"
	FreshnessAging   FreshnessStatus = "aging"
	FreshnessStale   FreshnessStatus = "stale"
	FreshnessUnknown FreshnessStatus = "unknown"
)

type Coordinate struct{ Longitude, Latitude float64 }

type FeedVersion struct {
	ID            int64
	SourceID      string
	VersionLabel  string
	ArchiveSHA256 string
	ActivatedAt   time.Time
}

type SourceHealth struct {
	SourceID            string
	LastAttemptAt       *time.Time
	LastSuccessAt       *time.Time
	LastValidSnapshotAt *time.Time
	ConsecutiveFailures int32
	LastErrorCode       *string
	EntityCount         *int32
}

type Vehicle struct {
	ID              string
	SourceID        string
	SourceVehicleID string
	RouteID         *string
	TripID          *string
	Mode            string
	Coordinate      Coordinate
	SourceUpdatedAt *time.Time
	EntityUpdatedAt *time.Time
	FetchedAt       time.Time
	ProcessedAt     time.Time
	Freshness       FreshnessStatus
	SnapshotID      int64
}

type NearbyStop struct {
	ID             string
	Name           string
	Mode           string
	Coordinate     Coordinate
	DistanceMeters float64
}

type VehicleFilter struct{ SourceIDs []string }
type NearbyStopsFilter struct {
	FeedVersionID int64
	Coordinate    Coordinate
	RadiusMeters  int32
	LimitPerMode  int32
	TotalLimit    int32
}

// Reader is safe for HTTP handlers: it contains no transaction ownership and
// returns normalized data only. Writers own their transaction boundaries.
type Reader interface {
	ActiveFeedVersion(ctx context.Context, sourceID string) (FeedVersion, error)
	SourceHealth(ctx context.Context, sourceID string) (SourceHealth, error)
	ListCurrentVehicles(ctx context.Context, filter VehicleFilter) ([]Vehicle, error)
	ListNearbyStops(ctx context.Context, filter NearbyStopsFilter) ([]NearbyStop, error)
}

// StaticImporterWriter supports a staging import followed by an atomic feed
// activation. Its implementation belongs to WP-04; callers must not mutate an
// active feed in place.
type StaticImporterWriter interface {
	CreateStagedFeed(ctx context.Context, sourceID, versionLabel, archiveSHA256 string, fetchedAt time.Time) (FeedVersion, error)
	ActivateFeed(ctx context.Context, sourceID string, feedVersionID int64, activatedAt time.Time) error
}

// RealtimeWriter makes valid snapshot replacement atomic. Invalid or empty
// upstream data must call RecordSourceFailure rather than delete current rows.
type RealtimeWriter interface {
	ReplaceVehicleSnapshot(ctx context.Context, snapshot VehicleSnapshot) error
	RecordSourceFailure(ctx context.Context, sourceID, safeCode string, attemptedAt time.Time) error
}

type VehicleSnapshot struct {
	SourceID        string
	SourceUpdatedAt *time.Time
	FetchedAt       time.Time
	ProcessedAt     time.Time
	ContentSHA256   string
	Vehicles        []Vehicle
}

func ValidatePublicID(id, kind string) error {
	prefix := ":" + kind + ":"
	if len(id) == 0 || len(id) > 512 || strings.ContainsAny(id, "\r\n\t") || !strings.Contains(id, prefix) {
		return ErrInvalidPublicID
	}
	parts := strings.Split(id, ":")
	if len(parts) < 3 || parts[0] == "" || parts[2] == "" {
		return ErrInvalidPublicID
	}
	return nil
}

// ServiceSeconds keeps GTFS time values such as 25:15:30 as seconds since
// service-day midnight. Conversion to an instant needs a service date/timezone.
func ServiceSeconds(hour, minute, second int) (int, error) {
	if hour < 0 || minute < 0 || minute > 59 || second < 0 || second > 59 {
		return 0, errors.New("invalid GTFS service-day time")
	}
	return hour*3600 + minute*60 + second, nil
}
