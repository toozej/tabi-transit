package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/toozej/tabi-transit/internal/sources/trimet"
)

// TriMetPlanner adapts the private TriMet Web Services DTO boundary to the
// provider-neutral application planning gateway. It deliberately maps only
// the normalized itinerary summary, never a provider payload or geometry.
type TriMetPlanner struct{ source trimet.Planner }

func NewTriMetPlanner(source trimet.Planner) TriMetPlanner { return TriMetPlanner{source: source} }

func (p TriMetPlanner) Plan(ctx context.Context, request JourneyRequest) (JourneyPlan, error) {
	if p.source == nil {
		return JourneyPlan{}, ErrUnavailable
	}
	origin, err := trimetPlace(request.Origin)
	if err != nil {
		return JourneyPlan{}, err
	}
	destination, err := trimetPlace(request.Destination)
	if err != nil {
		return JourneyPlan{}, err
	}
	providerRequest := trimet.PlanRequest{
		Origin: origin, Destination: destination,
		Preferences: trimet.PlanPreferences{
			Modes: trimetModes(request.Preferences.Modes), MaxTransfers: request.Preferences.MaxTransfers,
			MaxWalkMeters: request.Preferences.MaxWalkMeters, RequireAccessibility: request.Preferences.WheelchairAccessible,
		},
	}
	if request.Time != nil {
		providerRequest.DepartAt = &request.Time.Value
		providerRequest.ArriveBy = request.Time.Mode == JourneyArriveBy
	}
	plan, freshness, err := p.source.Plan(ctx, providerRequest)
	if err != nil {
		return JourneyPlan{}, err
	}
	out := JourneyPlan{PlanID: plan.ID, Provider: trimet.SourceID, Freshness: PlannerFreshness{
		Source: freshness.Source, FetchedAt: freshness.FetchedAt, ProcessedAt: freshness.ProcessedAt, IsRealtime: freshness.IsRealtime,
	}, Itineraries: make([]Itinerary, 0, len(plan.Itineraries))}
	for index, itinerary := range plan.Itineraries {
		item := Itinerary{ID: fmt.Sprintf("%s:%d", plan.ID, index+1), DurationSeconds: itinerary.DurationSeconds, Transfers: itinerary.Transfers, Legs: make([]JourneyLeg, 0, len(itinerary.Legs))}
		for _, leg := range itinerary.Legs {
			item.Legs = append(item.Legs, JourneyLeg{Mode: string(leg.Mode), RouteID: leg.RouteID, FromName: leg.FromName, ToName: leg.ToName, StartAt: leg.StartAt, EndAt: leg.EndAt, DistanceMeters: leg.DistanceMeters})
			if leg.DistanceMeters != nil && leg.Mode == trimet.ModeWalk {
				item.WalkMeters += *leg.DistanceMeters
			}
		}
		out.Itineraries = append(out.Itineraries, item)
	}
	return out, nil
}

func trimetPlace(place PlaceReference) (string, error) {
	switch place.Kind {
	case PlaceCoordinate, PlaceMapPin:
		if place.Coordinate == nil {
			return "", fmt.Errorf("invalid place reference")
		}
		return fmt.Sprintf("%.6f,%.6f", place.Coordinate.Latitude, place.Coordinate.Longitude), nil
	case PlaceStop, PlacePlace:
		id := strings.TrimSpace(place.ID)
		if id == "" {
			return "", fmt.Errorf("invalid place reference")
		}
		return strings.TrimPrefix(id, "trimet:stop:"), nil
	default:
		return "", fmt.Errorf("invalid place reference")
	}
}

func trimetModes(modes []string) []trimet.Mode {
	out := make([]trimet.Mode, 0, len(modes))
	for _, mode := range modes {
		switch mode {
		case "bus":
			out = append(out, trimet.ModeBus)
		case "light_rail", "commuter_rail":
			out = append(out, trimet.ModeRail)
		case "streetcar":
			out = append(out, trimet.ModeTram)
		case "walk":
			out = append(out, trimet.ModeWalk)
		}
	}
	return out
}
