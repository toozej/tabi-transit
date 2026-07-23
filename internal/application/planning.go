package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/toozej/tabi-transit/internal/persistence"
)

// ErrFeatureDisabled is deliberately distinct from ErrUnavailable. A disabled
// planning or search integration must not be presented as a temporary outage.
var ErrFeatureDisabled = errors.New("feature disabled")

// FeatureDisabledError contains only a stable feature/reason pair. In
// particular, it never includes a rider's search text, coordinates, or an
// upstream response.
type FeatureDisabledError struct {
	Feature string
	Reason  string
}

func (e *FeatureDisabledError) Error() string {
	return fmt.Sprintf("%s is disabled: %s", e.Feature, e.Reason)
}
func (e *FeatureDisabledError) Unwrap() error { return ErrFeatureDisabled }

const (
	FeatureJourneyPlanner = "journey_planner"
	FeaturePlaceSearch    = "place_search"
	ReasonExternalGate    = "external_provider_gate_pending"
)

// PlanningGateway is the provider-neutral seam used by a future composition
// root. It intentionally has no HTTP, credential, or provider DTO dependency.
// The nil/default gateway fails closed below.
type PlanningGateway interface {
	Plan(context.Context, JourneyRequest) (JourneyPlan, error)
}

// PlaceSearchGateway is intentionally separate from planning because Mapbox
// search/geocoding terms (D-004) may be approved independently of a planner.
type PlaceSearchGateway interface {
	Search(context.Context, PlaceSearchRequest) ([]PlaceResult, error)
}

type PlanningFeatures struct {
	Planner PlanningGateway
	Search  PlaceSearchGateway
}

type PlaceKind string

const (
	PlaceCoordinate PlaceKind = "coordinate"
	PlaceStop       PlaceKind = "stop"
	PlacePlace      PlaceKind = "place"
	PlaceMapPin     PlaceKind = "map_pin"
)

type PlaceReference struct {
	Kind       PlaceKind
	ID         string
	Coordinate *persistence.Coordinate
	// Label is display-only. Callers must never log it because it may be a
	// rider-supplied address or saved-place name.
	Label string
}

type JourneyTimeMode string

const (
	JourneyDepartAt JourneyTimeMode = "depart_at"
	JourneyArriveBy JourneyTimeMode = "arrive_by"
)

type JourneyTime struct {
	Mode  JourneyTimeMode
	Value time.Time
}

type JourneyPreferences struct {
	Modes                []string
	MaxTransfers         *int
	MaxWalkMeters        *int
	WheelchairAccessible bool
	Optimize             string
}

type JourneyRequest struct {
	Origin, Destination PlaceReference
	Time                *JourneyTime
	Preferences         JourneyPreferences
}

type JourneyPlan struct {
	PlanID                 string
	Provider               string
	ExpiresAt              *time.Time
	Itineraries            []Itinerary
	AppliedPreferences     []string
	UnsupportedPreferences []string
	Freshness              PlannerFreshness
}

// PlannerFreshness stays separate from persistence freshness because an
// itinerary is ephemeral and may have no normalized snapshot timestamp.
type PlannerFreshness struct {
	Source      string
	FetchedAt   time.Time
	ProcessedAt time.Time
	IsRealtime  bool
}

type Itinerary struct {
	ID              string
	DurationSeconds int
	Transfers       int
	WalkMeters      int
	Accessible      *bool
	Legs            []JourneyLeg
}

type JourneyLeg struct {
	Mode                      string
	RouteID, FromName, ToName string
	StartAt, EndAt            *time.Time
	DistanceMeters            *int
}

type PlaceSearchRequest struct {
	Query       string
	Types       []string
	Proximity   *persistence.Coordinate
	Language    string
	SessionHint string
}

type PlaceResult struct {
	ID          string
	Source      string
	Kind        string
	Name        string
	Subtitle    string
	Coordinate  *persistence.Coordinate
	StopID      string
	RouteID     string
	VehicleID   string
	Attribution string
}

func (s Service) PlanJourney(ctx context.Context, request JourneyRequest) (JourneyPlan, error) {
	if err := validateJourneyRequest(request); err != nil {
		return JourneyPlan{}, err
	}
	if s.Planning.Planner == nil {
		return JourneyPlan{}, disabled(FeatureJourneyPlanner)
	}
	plan, err := s.Planning.Planner.Plan(ctx, request)
	if err != nil {
		return JourneyPlan{}, err
	}
	plan.Itineraries = applyJourneyPolicy(plan.Itineraries, request.Preferences)
	return plan, nil
}

func (s Service) SearchPlaces(ctx context.Context, request PlaceSearchRequest) ([]PlaceResult, error) {
	if err := validatePlaceSearchRequest(request); err != nil {
		return nil, err
	}
	if s.Planning.Search == nil {
		return nil, disabled(FeaturePlaceSearch)
	}
	results, err := s.Planning.Search.Search(ctx, request)
	if err != nil {
		return nil, err
	}
	// Transit entities are always shown before external place results. Exact ID
	// matches are ordered first without retaining the original query anywhere.
	query := strings.ToLower(strings.TrimSpace(request.Query))
	sort.SliceStable(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if (a.Source == "tabi") != (b.Source == "tabi") {
			return a.Source == "tabi"
		}
		aExact := strings.EqualFold(a.ID, query) || strings.EqualFold(a.StopID, query) || strings.EqualFold(a.RouteID, query) || strings.EqualFold(a.VehicleID, query)
		bExact := strings.EqualFold(b.ID, query) || strings.EqualFold(b.StopID, query) || strings.EqualFold(b.RouteID, query) || strings.EqualFold(b.VehicleID, query)
		return aExact && !bExact
	})
	return results, nil
}

func disabled(feature string) error {
	return &FeatureDisabledError{Feature: feature, Reason: ReasonExternalGate}
}

func validateJourneyRequest(request JourneyRequest) error {
	if err := validatePlaceReference(request.Origin); err != nil {
		return err
	}
	if err := validatePlaceReference(request.Destination); err != nil {
		return err
	}
	if request.Time != nil && (request.Time.Value.IsZero() || (request.Time.Mode != JourneyDepartAt && request.Time.Mode != JourneyArriveBy)) {
		return errors.New("invalid journey time")
	}
	return validateJourneyPreferences(request.Preferences)
}

func validatePlaceReference(place PlaceReference) error {
	if len(place.ID) > 512 || len(place.Label) > 160 || containsControl(place.ID) || containsControl(place.Label) {
		return errors.New("invalid place reference")
	}
	switch place.Kind {
	case PlaceCoordinate, PlaceMapPin:
		if place.Coordinate == nil || !validCoordinate(*place.Coordinate) {
			return errors.New("invalid place reference")
		}
	case PlaceStop, PlacePlace:
		if strings.TrimSpace(place.ID) == "" {
			return errors.New("invalid place reference")
		}
	default:
		return errors.New("invalid place reference")
	}
	return nil
}

func validateJourneyPreferences(p JourneyPreferences) error {
	seen := make(map[string]struct{}, len(p.Modes))
	for _, mode := range p.Modes {
		switch mode {
		case "bus", "light_rail", "commuter_rail", "streetcar", "aerial_tram", "walk":
		default:
			return errors.New("invalid journey preferences")
		}
		if _, exists := seen[mode]; exists {
			return errors.New("invalid journey preferences")
		}
		seen[mode] = struct{}{}
	}
	if p.MaxTransfers != nil && (*p.MaxTransfers < 0 || *p.MaxTransfers > 5) {
		return errors.New("invalid journey preferences")
	}
	if p.MaxWalkMeters != nil && (*p.MaxWalkMeters < 0 || *p.MaxWalkMeters > 20000) {
		return errors.New("invalid journey preferences")
	}
	switch p.Optimize {
	case "", "fastest", "fewer_transfers", "least_walking":
	default:
		return errors.New("invalid journey preferences")
	}
	return nil
}

func validatePlaceSearchRequest(request PlaceSearchRequest) error {
	if containsControl(request.Query) {
		return errors.New("invalid place search request")
	}
	query := strings.TrimSpace(request.Query)
	if query == "" || len(query) > 160 || len(request.Language) > 35 || containsControl(request.Language) || len(request.SessionHint) > 128 || containsControl(request.SessionHint) {
		return errors.New("invalid place search request")
	}
	if request.Proximity != nil && !validCoordinate(*request.Proximity) {
		return errors.New("invalid place search request")
	}
	seen := make(map[string]struct{}, len(request.Types))
	for _, kind := range request.Types {
		switch kind {
		case "stop", "station", "route", "vehicle", "address", "poi", "intersection", "place":
		default:
			return errors.New("invalid place search request")
		}
		if _, exists := seen[kind]; exists {
			return errors.New("invalid place search request")
		}
		seen[kind] = struct{}{}
	}
	return nil
}

func applyJourneyPolicy(items []Itinerary, preferences JourneyPreferences) []Itinerary {
	allowedModes := make(map[string]struct{}, len(preferences.Modes))
	for _, mode := range preferences.Modes {
		allowedModes[mode] = struct{}{}
	}
	filtered := make([]Itinerary, 0, len(items))
	for _, item := range items {
		if item.DurationSeconds < 0 || item.Transfers < 0 || item.WalkMeters < 0 || (preferences.MaxTransfers != nil && item.Transfers > *preferences.MaxTransfers) || (preferences.MaxWalkMeters != nil && item.WalkMeters > *preferences.MaxWalkMeters) || (preferences.WheelchairAccessible && (item.Accessible == nil || !*item.Accessible)) {
			continue
		}
		matchesModes := true
		if len(allowedModes) > 0 {
			for _, leg := range item.Legs {
				if _, ok := allowedModes[leg.Mode]; !ok {
					matchesModes = false
					break
				}
			}
		}
		if matchesModes {
			filtered = append(filtered, item)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		switch preferences.Optimize {
		case "fewer_transfers":
			if a.Transfers != b.Transfers {
				return a.Transfers < b.Transfers
			}
		case "least_walking":
			if a.WalkMeters != b.WalkMeters {
				return a.WalkMeters < b.WalkMeters
			}
		}
		if a.DurationSeconds != b.DurationSeconds {
			return a.DurationSeconds < b.DurationSeconds
		}
		return a.ID < b.ID
	})
	return filtered
}

func validCoordinate(value persistence.Coordinate) bool {
	return value.Longitude >= -180 && value.Longitude <= 180 && value.Latitude >= -90 && value.Latitude <= 90
}
func containsControl(value string) bool { return strings.ContainsAny(value, "\r\n\t") }
