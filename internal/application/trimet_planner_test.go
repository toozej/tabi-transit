package application

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/toozej/tabi-transit/internal/persistence"
	"github.com/toozej/tabi-transit/internal/sources/trimet"
)

type fakeTriMetPlanner struct {
	request trimet.PlanRequest
	plan    trimet.Plan
	current trimet.Freshness
}

func (f *fakeTriMetPlanner) Plan(_ context.Context, request trimet.PlanRequest) (trimet.Plan, trimet.Freshness, error) {
	f.request = request
	return f.plan, f.current, nil
}

func TestTriMetPlannerMapsNormalizedRequestsAndResponses(t *testing.T) {
	t.Parallel()
	depart := time.Date(2026, 7, 28, 18, 30, 0, 0, time.UTC)
	fetched := depart.Add(time.Second)
	walk := 120
	source := &fakeTriMetPlanner{plan: trimet.Plan{ID: "provider-plan", Itineraries: []trimet.Itinerary{{DurationSeconds: 900, Transfers: 1, Legs: []trimet.ItineraryLeg{{Mode: trimet.ModeWalk, DistanceMeters: &walk}, {Mode: trimet.ModeBus, RouteID: "20"}}}}}, current: trimet.Freshness{Source: trimet.SourceID, FetchedAt: fetched, ProcessedAt: fetched, IsRealtime: true}}
	maximum := 2
	result, err := NewTriMetPlanner(source).Plan(context.Background(), JourneyRequest{
		Origin:      PlaceReference{Kind: PlaceCoordinate, Coordinate: coordinate(-122.67, 45.52)},
		Destination: PlaceReference{Kind: PlaceStop, ID: "trimet:stop:8334"},
		Time:        &JourneyTime{Mode: JourneyArriveBy, Value: depart},
		Preferences: JourneyPreferences{Modes: []string{"walk", "bus", "streetcar"}, MaxTransfers: &maximum, WheelchairAccessible: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if source.request.Origin != "45.520000,-122.670000" || source.request.Destination != "8334" || !source.request.ArriveBy || source.request.DepartAt == nil || !source.request.DepartAt.Equal(depart) {
		t.Fatalf("source request = %#v", source.request)
	}
	if got, want := source.request.Preferences.Modes, []trimet.Mode{trimet.ModeWalk, trimet.ModeBus, trimet.ModeTram}; !reflect.DeepEqual(got, want) {
		t.Fatalf("modes = %v, want %v", got, want)
	}
	if result.PlanID != "provider-plan" || result.Provider != trimet.SourceID || result.Freshness.Source != trimet.SourceID || len(result.Itineraries) != 1 || result.Itineraries[0].ID != "provider-plan:1" || result.Itineraries[0].WalkMeters != walk {
		t.Fatalf("result = %#v", result)
	}
}

func TestTriMetPlannerUnavailableWhenUncomposed(t *testing.T) {
	t.Parallel()
	_, err := NewTriMetPlanner(nil).Plan(context.Background(), planningRequest())
	if err != ErrUnavailable {
		t.Fatalf("error = %v, want %v", err, ErrUnavailable)
	}
}

func coordinate(longitude, latitude float64) *persistence.Coordinate {
	return &persistence.Coordinate{Longitude: longitude, Latitude: latitude}
}
