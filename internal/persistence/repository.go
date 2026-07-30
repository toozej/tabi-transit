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
var ErrNotFound = errors.New("normalized record not found")

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
type StaticFreshness struct {
	Source      string
	ActivatedAt time.Time
}
type StopDetail struct {
	Stop              CatalogStop
	StaticFeedVersion string
	Freshness         StaticFreshness
}
type RouteDetail struct {
	Route             CatalogRoute
	Directions        []RouteDirection
	StaticFeedVersion string
	Freshness         StaticFreshness
}
type RouteDirection struct {
	ID       int
	Headsign string
}
type RouteStop struct {
	Stop        CatalogStop
	Sequence    int
	DirectionID *int
}
type RouteShape struct {
	ID, RouteID string
	DirectionID *int
	Coordinates [][]float64
}

// ScheduleTime intentionally retains GTFS service-day seconds. Values past
// 86400 describe after-midnight service and must not be wrapped to midnight.
type ScheduleTime struct {
	TripID, RouteID, StopID, ServiceDate string
	DirectionID                          *int
	Headsign                             string
	ServiceDaySeconds                    int
	DepartureAt                          time.Time
}
type Arrival struct {
	ID, StopID, RouteID, Status string
	DirectionID                 *int
	Headsign                    string
	ScheduledAt                 time.Time
	EstimatedAt                 *time.Time
	TripID                      *string
	StopSequence                *int
	Freshness                   StaticFreshness
}
type Alert struct {
	ID, Revision, Header, Description, Cause, Effect, SourceURL, Source string
	ActiveFrom, ActiveUntil                                             *time.Time
	FetchedAt, ProcessedAt                                              time.Time
}
type ScheduleFilter struct {
	StopID, ServiceDate, RouteID, Cursor string
	DirectionID                          *int
	Limit                                int
}
type ArrivalFilter struct {
	StopID           string
	Minutes          int
	RouteIDs         []string
	DirectionID      *int
	IncludeScheduled bool
	Now              time.Time
}
type AlertFilter struct {
	RouteIDs, StopIDs, Modes []string
	Effect                   string
	Active                   bool
	UpdatedSince             *time.Time
	Cursor                   string
	Limit                    int
	Now                      time.Time
}

type VehicleFilter struct{ SourceIDs []string }

// VehicleHistoryFilter is deliberately bounded by the caller to the approved
// retention interval. Cursor is an exclusive processed-at keyset cursor.
type VehicleHistoryFilter struct {
	VehicleID string
	From, To  time.Time
	Limit     int
	Cursor    *time.Time
}
type VehicleObservation struct {
	VehicleID, SourceID, SourceVehicleID string
	RouteID, TripID                      *string
	Mode                                 string
	Coordinate                           Coordinate
	ObservedAt, FetchedAt                time.Time
	Freshness                            FreshnessStatus
}
type CatalogRoute struct {
	ID, Mode, ShortName, LongName string
	Color, TextColor              *string
}
type CatalogStop struct {
	ID, Name             string
	Coordinate           Coordinate
	Modes, RouteIDs      []string
	ParentStopID         *string
	WheelchairAccessible *bool
}
type CatalogFilter struct {
	Modes         []string
	Query, Cursor string
	Limit         int
}
type CatalogPage[T any] struct {
	Items             []T
	NextCursor        string
	StaticFeedVersion string
}
type NearbyStopsFilter struct {
	Coordinate           Coordinate
	RadiusMeters         int32
	LimitPerMode         int32
	TotalLimit           int32
	Modes                []string
	WheelchairAccessible *bool
}

// Reader is safe for HTTP handlers: it contains no transaction ownership and
// returns normalized data only. Writers own their transaction boundaries.
type Reader interface {
	ActiveFeedVersion(ctx context.Context, sourceID string) (FeedVersion, error)
	SourceHealth(ctx context.Context, sourceID string) (SourceHealth, error)
	ListCurrentVehicles(ctx context.Context, filter VehicleFilter) ([]Vehicle, error)
	ListVehicleHistory(ctx context.Context, filter VehicleHistoryFilter) ([]VehicleObservation, error)
	ListNearbyStops(ctx context.Context, filter NearbyStopsFilter) ([]NearbyStop, error)
}

// CatalogReader keeps static-feed SQL details out of application handlers. Its
// cursor is a storage key; public cursor encoding remains in application.
type CatalogReader interface {
	ListCatalogRoutes(ctx context.Context, filter CatalogFilter) (CatalogPage[CatalogRoute], error)
	ListCatalogStops(ctx context.Context, filter CatalogFilter) (CatalogPage[CatalogStop], error)
}

// RiderInfoReader exposes only static GTFS material that the current schema
// can prove. Arrivals remain unavailable until normalized trip-update storage
// and a service calendar are implemented.
type RiderInfoReader interface {
	ListNearbyStops(context.Context, NearbyStopsFilter) ([]NearbyStop, error)
	GetStop(context.Context, string) (StopDetail, error)
	GetRoute(context.Context, string) (RouteDetail, error)
	ListRouteStops(context.Context, string, *int) ([]RouteStop, string, error)
	ListRouteShapes(context.Context, string, *int) ([]RouteShape, string, error)
	ListStopSchedule(context.Context, ScheduleFilter) ([]ScheduleTime, string, string, error)
	ListStopArrivals(context.Context, ArrivalFilter) ([]Arrival, error)
	ListAlerts(context.Context, AlertFilter) ([]Alert, string, error)
	GetAlert(context.Context, string) (Alert, error)
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
	ReplaceTripUpdateSnapshot(ctx context.Context, snapshot TripUpdateSnapshot) error
	RecordSourceFailure(ctx context.Context, sourceID, safeCode string, attemptedAt time.Time) error
}

// TripUpdateSnapshot is a complete validated feed projection. An empty or
// malformed provider response is never represented here, so writers can
// safely replace current rows atomically.
type TripUpdateSnapshot struct {
	SourceID        string
	SourceUpdatedAt *time.Time
	FetchedAt       time.Time
	ProcessedAt     time.Time
	ContentSHA256   string
	Updates         []TripUpdate
}
type TripUpdate struct {
	EntityID, TripID, RouteID, StartDate, ScheduleRelationship string
	UpdatedAt                                                  *time.Time
	StopTimes                                                  []TripUpdateStopTime
}
type TripUpdateStopTime struct {
	StopSequence                               int
	StopID                                     string
	ArrivalDelaySeconds, DepartureDelaySeconds *int32
	ArrivalTime, DepartureTime                 *time.Time
	ScheduleRelationship                       string
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
