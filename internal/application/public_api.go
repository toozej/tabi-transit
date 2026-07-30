// Package application holds use cases independent of HTTP and database drivers.
package application

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/toozej/tabi-transit/internal/persistence"
)

var ErrUnavailable = errors.New("dependency unavailable")

type Route struct {
	ID, Mode, ShortName, LongName string
	Color, TextColor              *string
}
type Stop struct {
	ID, Name             string
	Coordinate           persistence.Coordinate
	Modes, RouteIDs      []string
	ParentStopID         *string
	WheelchairAccessible *bool
}
type Page[T any] struct {
	Items             []T
	NextCursor        string
	StaticFeedVersion string
}
type Catalog interface {
	ListRoutes(context.Context, RouteQuery) (Page[Route], error)
	ListStops(context.Context, StopQuery) (Page[Stop], error)
}
type RouteQuery struct {
	Modes              []string
	Query, ServiceDate string
	Limit              int
	Cursor             string
}
type StopQuery struct {
	Modes  []string
	Query  string
	Limit  int
	Cursor string
}
type VehicleStore interface {
	ListCurrentVehicles(context.Context, persistence.VehicleFilter) ([]persistence.Vehicle, error)
}
type VehicleHistoryStore interface {
	ListVehicleHistory(context.Context, persistence.VehicleHistoryFilter) ([]persistence.VehicleObservation, error)
}
type Service struct {
	Catalog   Catalog
	Vehicles  VehicleStore
	History   VehicleHistoryStore
	RiderInfo RiderInfo
	// Planning remains feature-gated until D-001 (TriMet planner) and D-004
	// (Mapbox search/geocoding) have reviewed credentials, terms, retention,
	// attribution, and supported constraint semantics.
	Planning PlanningFeatures
	// Notifications stays disabled until encrypted persistence and the Expo
	// Push decision/credentials are explicitly composed by the runtime.
	Notifications NotificationFeatures
	Now           func() time.Time
}

type RiderInfo interface {
	NearbyStops(context.Context, NearbyQuery) ([]persistence.NearbyStop, error)
	Stop(context.Context, string) (persistence.StopDetail, error)
	Route(context.Context, string) (persistence.RouteDetail, error)
	RouteStops(context.Context, string, *int) ([]persistence.RouteStop, string, error)
	RouteShapes(context.Context, string, *int) ([]persistence.RouteShape, string, error)
	StopSchedule(context.Context, persistence.ScheduleFilter) ([]persistence.ScheduleTime, string, string, error)
	StopArrivals(context.Context, persistence.ArrivalFilter) ([]persistence.Arrival, error)
	Alerts(context.Context, persistence.AlertFilter) ([]persistence.Alert, string, error)
	Alert(context.Context, string) (persistence.Alert, error)
}

func (s Service) StopSchedule(ctx context.Context, q persistence.ScheduleFilter) ([]persistence.ScheduleTime, string, string, error) {
	if s.RiderInfo == nil {
		return nil, "", "", ErrUnavailable
	}
	return s.RiderInfo.StopSchedule(ctx, q)
}
func (s Service) StopArrivals(ctx context.Context, q persistence.ArrivalFilter) ([]persistence.Arrival, error) {
	if s.RiderInfo == nil {
		return nil, ErrUnavailable
	}
	return s.RiderInfo.StopArrivals(ctx, q)
}
func (s Service) Alerts(ctx context.Context, q persistence.AlertFilter) ([]persistence.Alert, string, error) {
	if s.RiderInfo == nil {
		return nil, "", ErrUnavailable
	}
	return s.RiderInfo.Alerts(ctx, q)
}
func (s Service) Alert(ctx context.Context, id string) (persistence.Alert, error) {
	if s.RiderInfo == nil {
		return persistence.Alert{}, ErrUnavailable
	}
	return s.RiderInfo.Alert(ctx, id)
}

type NearbyQuery struct {
	Coordinate                        persistence.Coordinate
	RadiusMeters, LimitPerMode, Limit int
	Modes                             []string
	WheelchairAccessible              *bool
}

func (s Service) ListRoutes(ctx context.Context, q RouteQuery) (Page[Route], error) {
	if s.Catalog == nil {
		return Page[Route]{}, ErrUnavailable
	}
	return s.Catalog.ListRoutes(ctx, q)
}
func (s Service) ListStops(ctx context.Context, q StopQuery) (Page[Stop], error) {
	if s.Catalog == nil {
		return Page[Stop]{}, ErrUnavailable
	}
	return s.Catalog.ListStops(ctx, q)
}
func (s Service) ListVehicles(ctx context.Context, sourceIDs []string) ([]persistence.Vehicle, error) {
	if s.Vehicles == nil {
		return nil, ErrUnavailable
	}
	return s.Vehicles.ListCurrentVehicles(ctx, persistence.VehicleFilter{SourceIDs: sourceIDs})
}
func (s Service) SearchVehicles(ctx context.Context, query string, limit int) ([]persistence.Vehicle, error) {
	vs, err := s.ListVehicles(ctx, nil)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	out := make([]persistence.Vehicle, 0, limit)
	for _, v := range vs {
		if strings.Contains(strings.ToLower(v.ID), q) || strings.Contains(strings.ToLower(v.SourceVehicleID), q) {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		ae := a.ID == query || a.SourceVehicleID == query
		be := b.ID == query || b.SourceVehicleID == query
		return ae && !be
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (s Service) Vehicle(ctx context.Context, id string) (persistence.Vehicle, error) {
	vs, err := s.ListVehicles(ctx, nil)
	if err != nil {
		return persistence.Vehicle{}, err
	}
	for _, v := range vs {
		if v.ID == id {
			return v, nil
		}
	}
	return persistence.Vehicle{}, errors.New("not found")
}
func (s Service) VehicleHistory(ctx context.Context, q persistence.VehicleHistoryFilter) ([]persistence.VehicleObservation, error) {
	if s.History == nil {
		return nil, ErrUnavailable
	}
	return s.History.ListVehicleHistory(ctx, q)
}
func (s Service) NearbyStops(ctx context.Context, q NearbyQuery) ([]persistence.NearbyStop, error) {
	if s.RiderInfo == nil {
		return nil, ErrUnavailable
	}
	return s.RiderInfo.NearbyStops(ctx, q)
}
func (s Service) Stop(ctx context.Context, id string) (persistence.StopDetail, error) {
	if s.RiderInfo == nil {
		return persistence.StopDetail{}, ErrUnavailable
	}
	return s.RiderInfo.Stop(ctx, id)
}
func (s Service) Route(ctx context.Context, id string) (persistence.RouteDetail, error) {
	if s.RiderInfo == nil {
		return persistence.RouteDetail{}, ErrUnavailable
	}
	return s.RiderInfo.Route(ctx, id)
}
func (s Service) RouteStops(ctx context.Context, id string, direction *int) ([]persistence.RouteStop, string, error) {
	if s.RiderInfo == nil {
		return nil, "", ErrUnavailable
	}
	return s.RiderInfo.RouteStops(ctx, id, direction)
}
func (s Service) RouteShapes(ctx context.Context, id string, direction *int) ([]persistence.RouteShape, string, error) {
	if s.RiderInfo == nil {
		return nil, "", ErrUnavailable
	}
	return s.RiderInfo.RouteShapes(ctx, id, direction)
}
