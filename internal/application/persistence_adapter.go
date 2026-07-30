package application

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/toozej/tabi-transit/internal/persistence"
)

type PersistenceVehicleStore struct{ Reader persistence.Reader }

func (s PersistenceVehicleStore) ListCurrentVehicles(ctx context.Context, filter persistence.VehicleFilter) ([]persistence.Vehicle, error) {
	return s.Reader.ListCurrentVehicles(ctx, filter)
}
func (s PersistenceVehicleStore) ListVehicleHistory(ctx context.Context, filter persistence.VehicleHistoryFilter) ([]persistence.VehicleObservation, error) {
	return s.Reader.ListVehicleHistory(ctx, filter)
}

// PersistenceCatalog turns the public opaque cursor into the database keyset
// cursor and maps persistence records into the public application model.
type PersistenceCatalog struct{ Reader persistence.CatalogReader }
type PersistenceRiderInfo struct{ Reader persistence.RiderInfoReader }

func (s PersistenceRiderInfo) NearbyStops(ctx context.Context, q NearbyQuery) ([]persistence.NearbyStop, error) {
	return s.Reader.ListNearbyStops(ctx, persistence.NearbyStopsFilter{Coordinate: q.Coordinate, RadiusMeters: int32(q.RadiusMeters), LimitPerMode: int32(q.LimitPerMode), TotalLimit: int32(q.Limit), Modes: q.Modes, WheelchairAccessible: q.WheelchairAccessible})
}
func (s PersistenceRiderInfo) Stop(ctx context.Context, id string) (persistence.StopDetail, error) {
	return s.Reader.GetStop(ctx, id)
}
func (s PersistenceRiderInfo) Route(ctx context.Context, id string) (persistence.RouteDetail, error) {
	return s.Reader.GetRoute(ctx, id)
}
func (s PersistenceRiderInfo) RouteStops(ctx context.Context, id string, direction *int) ([]persistence.RouteStop, string, error) {
	return s.Reader.ListRouteStops(ctx, id, direction)
}
func (s PersistenceRiderInfo) RouteShapes(ctx context.Context, id string, direction *int) ([]persistence.RouteShape, string, error) {
	return s.Reader.ListRouteShapes(ctx, id, direction)
}
func (s PersistenceRiderInfo) StopSchedule(ctx context.Context, q persistence.ScheduleFilter) ([]persistence.ScheduleTime, string, string, error) {
	return s.Reader.ListStopSchedule(ctx, q)
}
func (s PersistenceRiderInfo) StopArrivals(ctx context.Context, q persistence.ArrivalFilter) ([]persistence.Arrival, error) {
	return s.Reader.ListStopArrivals(ctx, q)
}
func (s PersistenceRiderInfo) Alerts(ctx context.Context, q persistence.AlertFilter) ([]persistence.Alert, string, error) {
	return s.Reader.ListAlerts(ctx, q)
}
func (s PersistenceRiderInfo) Alert(ctx context.Context, id string) (persistence.Alert, error) {
	return s.Reader.GetAlert(ctx, id)
}

func (s PersistenceCatalog) ListRoutes(ctx context.Context, q RouteQuery) (Page[Route], error) {
	page, err := s.Reader.ListCatalogRoutes(ctx, persistence.CatalogFilter{
		Modes: q.Modes, Query: q.Query, Cursor: decodeCursor(q.Cursor), Limit: q.Limit,
	})
	if err != nil {
		return Page[Route]{}, err
	}
	items := make([]Route, len(page.Items))
	for i, route := range page.Items {
		items[i] = Route{ID: route.ID, Mode: route.Mode, ShortName: route.ShortName, LongName: route.LongName, Color: route.Color, TextColor: route.TextColor}
	}
	return Page[Route]{Items: items, NextCursor: encodeCursor(page.NextCursor), StaticFeedVersion: page.StaticFeedVersion}, nil
}

func (s PersistenceCatalog) ListStops(ctx context.Context, q StopQuery) (Page[Stop], error) {
	page, err := s.Reader.ListCatalogStops(ctx, persistence.CatalogFilter{
		Modes: q.Modes, Query: q.Query, Cursor: decodeCursor(q.Cursor), Limit: q.Limit,
	})
	if err != nil {
		return Page[Stop]{}, err
	}
	items := make([]Stop, len(page.Items))
	for i, stop := range page.Items {
		items[i] = Stop{ID: stop.ID, Name: stop.Name, Coordinate: stop.Coordinate, Modes: stop.Modes, RouteIDs: stop.RouteIDs, ParentStopID: stop.ParentStopID, WheelchairAccessible: stop.WheelchairAccessible}
	}
	return Page[Stop]{Items: items, NextCursor: encodeCursor(page.NextCursor), StaticFeedVersion: page.StaticFeedVersion}, nil
}

func encodeCursor(value string) string {
	if value == "" {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func decodeCursor(value string) string {
	if value == "" {
		return ""
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	decodedValue := string(decoded)
	if err != nil || len(decoded) == 0 || len(decoded) > 512 || strings.ContainsAny(decodedValue, "\r\n\t") || strings.Count(decodedValue, ":") < 2 {
		// API validation owns the public request. Treat an unrecognised legacy or
		// malformed cursor as an initial page rather than using it in SQL.
		return ""
	}
	return decodedValue
}
