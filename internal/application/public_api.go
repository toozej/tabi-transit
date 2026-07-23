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
type Service struct {
	Catalog  Catalog
	Vehicles VehicleStore
	Now      func() time.Time
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
