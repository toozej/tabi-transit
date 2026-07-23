package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/toozej/tabi-transit/internal/persistence"
)

type fixturePlanner struct {
	called bool
	plan   JourneyPlan
}

func (f *fixturePlanner) Plan(_ context.Context, _ JourneyRequest) (JourneyPlan, error) {
	f.called = true
	return f.plan, nil
}

type fixtureSearch struct {
	called bool
	items  []PlaceResult
}

func (f *fixtureSearch) Search(_ context.Context, _ PlaceSearchRequest) ([]PlaceResult, error) {
	f.called = true
	return f.items, nil
}

func planningRequest() JourneyRequest {
	return JourneyRequest{
		Origin:      PlaceReference{Kind: PlaceCoordinate, Coordinate: &persistence.Coordinate{Longitude: -122.67, Latitude: 45.52}},
		Destination: PlaceReference{Kind: PlaceStop, ID: "trimet:stop:8334"},
		Preferences: JourneyPreferences{Modes: []string{"walk", "bus"}, Optimize: "fewer_transfers"},
	}
}

func TestPlanningFeaturesFailClosed(t *testing.T) {
	t.Parallel()
	service := Service{}
	if _, err := service.PlanJourney(context.Background(), planningRequest()); !errors.Is(err, ErrFeatureDisabled) {
		t.Fatalf("planner error = %v", err)
	}
	if _, err := service.SearchPlaces(context.Background(), PlaceSearchRequest{Query: "8334"}); !errors.Is(err, ErrFeatureDisabled) {
		t.Fatalf("search error = %v", err)
	}
	var disabledErr *FeatureDisabledError
	if _, err := service.PlanJourney(context.Background(), planningRequest()); !errors.As(err, &disabledErr) || disabledErr.Feature != FeatureJourneyPlanner || disabledErr.Reason != ReasonExternalGate {
		t.Fatalf("unexpected disabled error: %v", err)
	}
}

func TestPlanJourneyValidatesBeforeGatewayAndAppliesHardConstraints(t *testing.T) {
	t.Parallel()
	accessible, inaccessible := true, false
	maxTransfers, maxWalk := 1, 900
	planner := &fixturePlanner{plan: JourneyPlan{Itineraries: []Itinerary{
		{ID: "too-many-transfers", DurationSeconds: 100, Transfers: 2, WalkMeters: 300, Accessible: &accessible, Legs: []JourneyLeg{{Mode: "bus"}}},
		{ID: "inaccessible", DurationSeconds: 100, Transfers: 0, WalkMeters: 300, Accessible: &inaccessible, Legs: []JourneyLeg{{Mode: "bus"}}},
		{ID: "wrong-mode", DurationSeconds: 100, Transfers: 0, WalkMeters: 300, Accessible: &accessible, Legs: []JourneyLeg{{Mode: "light_rail"}}},
		{ID: "slow", DurationSeconds: 500, Transfers: 1, WalkMeters: 800, Accessible: &accessible, Legs: []JourneyLeg{{Mode: "walk"}, {Mode: "bus"}}},
		{ID: "fast", DurationSeconds: 300, Transfers: 1, WalkMeters: 700, Accessible: &accessible, Legs: []JourneyLeg{{Mode: "bus"}}},
	}}}
	request := planningRequest()
	request.Preferences.MaxTransfers = &maxTransfers
	request.Preferences.MaxWalkMeters = &maxWalk
	request.Preferences.WheelchairAccessible = true
	result, err := (Service{Planning: PlanningFeatures{Planner: planner}}).PlanJourney(context.Background(), request)
	if err != nil || !planner.called {
		t.Fatalf("plan error = %v called=%t", err, planner.called)
	}
	if got := []string{result.Itineraries[0].ID, result.Itineraries[1].ID}; !reflect.DeepEqual(got, []string{"fast", "slow"}) {
		t.Fatalf("filtered/ranked IDs = %v", got)
	}

	planner.called = false
	bad := planningRequest()
	bad.Origin.Label = "secret address\n"
	if _, err := (Service{Planning: PlanningFeatures{Planner: planner}}).PlanJourney(context.Background(), bad); err == nil || planner.called || err.Error() != "invalid place reference" {
		t.Fatalf("unsafe input must fail before gateway: err=%v called=%t", err, planner.called)
	}
}

func TestSearchPlacesOrdersTransitAndExactIDsWithoutLeakingInput(t *testing.T) {
	t.Parallel()
	search := &fixtureSearch{items: []PlaceResult{
		{ID: "mapbox:place:one", Source: "mapbox", Kind: "place", Name: "Provider result"},
		{ID: "trimet:stop:8334", Source: "tabi", Kind: "stop", Name: "Stop"},
		{ID: "trimet:route:20", Source: "tabi", Kind: "route", Name: "Route"},
	}}
	result, err := (Service{Planning: PlanningFeatures{Search: search}}).SearchPlaces(context.Background(), PlaceSearchRequest{Query: "trimet:stop:8334"})
	if err != nil || !search.called {
		t.Fatalf("search error = %v called=%t", err, search.called)
	}
	if got := []string{result[0].ID, result[1].ID, result[2].ID}; !reflect.DeepEqual(got, []string{"trimet:stop:8334", "trimet:route:20", "mapbox:place:one"}) {
		t.Fatalf("search order = %v", got)
	}
	search.called = false
	if _, err := (Service{Planning: PlanningFeatures{Search: search}}).SearchPlaces(context.Background(), PlaceSearchRequest{Query: "private destination\n"}); err == nil || search.called || err.Error() != "invalid place search request" {
		t.Fatalf("unsafe search must fail safely: err=%v called=%t", err, search.called)
	}
}
