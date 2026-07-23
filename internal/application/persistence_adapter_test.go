package application

import (
	"context"
	"testing"

	"github.com/toozej/tabi-transit/internal/persistence"
)

type catalogReaderFixture struct {
	routeFilter persistence.CatalogFilter
	stopFilter  persistence.CatalogFilter
}

func (f *catalogReaderFixture) ListCatalogRoutes(_ context.Context, filter persistence.CatalogFilter) (persistence.CatalogPage[persistence.CatalogRoute], error) {
	f.routeFilter = filter
	return persistence.CatalogPage[persistence.CatalogRoute]{Items: []persistence.CatalogRoute{{ID: "fixture:route:20", Mode: "bus", ShortName: "20"}}, NextCursor: "fixture:route:20", StaticFeedVersion: "fixture-v1"}, nil
}
func (f *catalogReaderFixture) ListCatalogStops(_ context.Context, filter persistence.CatalogFilter) (persistence.CatalogPage[persistence.CatalogStop], error) {
	f.stopFilter = filter
	return persistence.CatalogPage[persistence.CatalogStop]{Items: []persistence.CatalogStop{{ID: "fixture:stop:1", Name: "Fixture", Modes: []string{"bus"}}}, NextCursor: "fixture:stop:1", StaticFeedVersion: "fixture-v1"}, nil
}

func TestPersistenceCatalogMapsOpaqueCursors(t *testing.T) {
	t.Parallel()
	reader := &catalogReaderFixture{}
	catalog := PersistenceCatalog{Reader: reader}
	routes, err := catalog.ListRoutes(context.Background(), RouteQuery{Modes: []string{"bus"}, Query: "20", Cursor: "Zml4dHVyZTpyb3V0ZTo5", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if reader.routeFilter.Cursor != "fixture:route:9" || routes.NextCursor != "Zml4dHVyZTpyb3V0ZToyMA" || routes.StaticFeedVersion != "fixture-v1" {
		t.Fatalf("route mapping mismatch: %#v %#v", reader.routeFilter, routes)
	}
	stops, err := catalog.ListStops(context.Background(), StopQuery{Query: "fixture", Cursor: "not-a-cursor", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if reader.stopFilter.Cursor != "" || stops.NextCursor != "Zml4dHVyZTpzdG9wOjE" {
		t.Fatalf("stop mapping mismatch: %#v %#v", reader.stopFilter, stops)
	}
}
