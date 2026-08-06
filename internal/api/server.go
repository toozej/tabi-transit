// Package api implements the public HTTP boundary. It never logs request query
// values or coordinates, because both can be sensitive rider data.
package api

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/toozej/tabi-transit/internal/application"
	"github.com/toozej/tabi-transit/internal/config"
	"github.com/toozej/tabi-transit/internal/persistence"
)

type Server struct {
	app     application.Service
	config  config.Config
	ready   func(context.Context) error
	limiter *rateLimiter
}
type Option func(*Server)

func WithReadiness(fn func(context.Context) error) Option { return func(s *Server) { s.ready = fn } }
func New(app application.Service, c config.Config, options ...Option) http.Handler {
	s := &Server{app: app, config: c, limiter: newRateLimiter(c.RateLimit)}
	for _, o := range options {
		o(s)
	}
	r := chi.NewRouter()
	r.Use(s.requestID, s.recover, s.cors, s.securityHeaders, s.limit)
	r.Get("/health/live", s.live)
	r.Get("/health/ready", s.readiness)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/config", s.getConfig)
		r.Get("/static/manifest", s.staticManifest)
		r.Post("/installations", s.createInstallation)
		r.Put("/installations/{id}/push-token", s.registerPushToken)
		r.Delete("/installations/{id}", s.deleteInstallation)
		r.Get("/subscriptions", s.listSubscriptions)
		r.Post("/subscriptions", s.createSubscription)
		r.Delete("/subscriptions/{id}", s.deleteSubscription)
		r.Get("/search", s.placeSearch)
		r.Get("/geocode/reverse", s.reverseGeocode)
		r.Post("/journeys/plan", s.planJourney)
		r.Get("/routes", s.routes)
		r.Get("/stops", s.stops)
		r.Get("/stops/nearby", s.nearbyStops)
		r.Get("/stops/{id}", s.stop)
		r.Get("/stops/{id}/arrivals", s.stopArrivals)
		r.Get("/stops/{id}/schedule", s.stopSchedule)
		r.Get("/routes/{id}", s.route)
		r.Get("/routes/{id}/shape", s.routeShape)
		r.Get("/routes/{id}/stops", s.routeStops)
		r.Get("/routes/{id}/vehicles", s.routeVehicles)
		r.Get("/alerts", s.alerts)
		r.Get("/alerts/{id}", s.alert)
		r.Get("/vehicles", s.vehicles)
		r.Get("/vehicles/search", s.vehicleSearch)
		r.Get("/vehicles/{id}/history", s.vehicleHistory)
		r.Get("/vehicles/{id}", s.vehicle)
	})
	return r
}

// staticManifest is deliberately artifact-free until artifact storage and
// retention are configured. A valid empty manifest still lets mobile clients
// synchronize the active static-feed version without receiving a 404.
func (s *Server) staticManifest(w http.ResponseWriter, r *http.Request) {
	published := s.config.API.StaticFeedPublishedAt.UTC()
	if published.IsZero() {
		published = time.Unix(0, 0).UTC()
	}
	w.Header().Set("X-Static-Feed-Version", s.config.API.StaticFeedVersion)
	s.etag(w, r, map[string]any{
		"staticFeedVersion": s.config.API.StaticFeedVersion,
		"publishedAt":       published.Format(time.RFC3339),
		"artifacts":         []any{},
		"freshness":         staticFreshness(published),
	})
}

// The external-place endpoints deliberately fail closed until their documented
// Mapbox decision gate is approved and an adapter is composed.
func (s *Server) placeSearch(w http.ResponseWriter, r *http.Request) {
	s.featureUnavailable(w, r, application.FeaturePlaceSearch)
}

func (s *Server) reverseGeocode(w http.ResponseWriter, r *http.Request) {
	s.featureUnavailable(w, r, application.FeaturePlaceSearch)
}

func (s *Server) planJourney(w http.ResponseWriter, r *http.Request) {
	if s.app.Planning.Planner == nil {
		s.featureUnavailable(w, r, application.FeatureJourneyPlanner)
		return
	}
	var input journeyPlanInput
	decoder := json.NewDecoder(io.LimitReader(r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		s.validation(w, r, "body", "invalid")
		return
	}
	request, err := input.applicationRequest()
	if err != nil {
		s.validation(w, r, "body", "invalid")
		return
	}
	plan, err := s.app.PlanJourney(r.Context(), request)
	if errors.Is(err, application.ErrFeatureDisabled) || errors.Is(err, application.ErrUnavailable) {
		s.featureUnavailable(w, r, application.FeatureJourneyPlanner)
		return
	}
	if err != nil {
		s.unavailable(w, r)
		return
	}
	s.write(w, http.StatusOK, journeyPlanJSON(request, plan))
}

type journeyPlanInput struct {
	Origin      journeyLocationInput    `json:"origin"`
	Destination journeyLocationInput    `json:"destination"`
	Time        journeyTimeInput        `json:"time"`
	Preferences journeyPreferencesInput `json:"preferences"`
}
type journeyLocationInput struct {
	Type       string    `json:"type"`
	Coordinate []float64 `json:"coordinate"`
	StopID     string    `json:"stopId"`
	PlaceID    string    `json:"placeId"`
	LocalID    string    `json:"localId"`
	Label      string    `json:"label"`
}
type journeyTimeInput struct {
	Mode  string    `json:"mode"`
	Value time.Time `json:"value"`
}
type journeyPreferencesInput struct {
	Modes                []string `json:"modes"`
	MaxTransfers         *int     `json:"maxTransfers"`
	MaxWalkMeters        *int     `json:"maxWalkingMeters"`
	WheelchairAccessible bool     `json:"wheelchairAccessible"`
	Optimize             string   `json:"optimize"`
}

func (input journeyPlanInput) applicationRequest() (application.JourneyRequest, error) {
	origin, err := input.Origin.applicationPlace()
	if err != nil {
		return application.JourneyRequest{}, err
	}
	destination, err := input.Destination.applicationPlace()
	if err != nil {
		return application.JourneyRequest{}, err
	}
	if input.Time.Value.IsZero() {
		return application.JourneyRequest{}, errors.New("time is required")
	}
	return application.JourneyRequest{Origin: origin, Destination: destination, Time: &application.JourneyTime{Mode: application.JourneyTimeMode(input.Time.Mode), Value: input.Time.Value}, Preferences: application.JourneyPreferences{Modes: input.Preferences.Modes, MaxTransfers: input.Preferences.MaxTransfers, MaxWalkMeters: input.Preferences.MaxWalkMeters, WheelchairAccessible: input.Preferences.WheelchairAccessible, Optimize: input.Preferences.Optimize}}, nil
}
func (input journeyLocationInput) applicationPlace() (application.PlaceReference, error) {
	result := application.PlaceReference{Label: input.Label}
	switch input.Type {
	case "coordinate":
		result.Kind = application.PlaceCoordinate
	case "stop":
		result.Kind, result.ID = application.PlaceStop, input.StopID
	case "place":
		result.Kind, result.ID = application.PlacePlace, input.PlaceID
	case "saved", "recent":
		result.Kind, result.ID = application.PlaceMapPin, input.LocalID
	default:
		return application.PlaceReference{}, errors.New("invalid place type")
	}
	if len(input.Coordinate) == 2 {
		result.Coordinate = &persistence.Coordinate{Longitude: input.Coordinate[0], Latitude: input.Coordinate[1]}
	}
	return result, nil
}
func journeyPlanJSON(request application.JourneyRequest, plan application.JourneyPlan) map[string]any {
	itineraries := make([]any, 0, len(plan.Itineraries))
	for _, itinerary := range plan.Itineraries {
		legs := make([]any, 0, len(itinerary.Legs))
		for _, leg := range itinerary.Legs {
			legJSON := map[string]any{"mode": leg.Mode, "fromName": leg.FromName, "toName": leg.ToName}
			if leg.RouteID != "" {
				legJSON["routeId"] = leg.RouteID
			}
			if leg.StartAt != nil {
				legJSON["startAt"] = leg.StartAt.UTC().Format(time.RFC3339)
			}
			if leg.EndAt != nil {
				legJSON["endAt"] = leg.EndAt.UTC().Format(time.RFC3339)
			}
			if leg.DistanceMeters != nil {
				legJSON["distanceMeters"] = *leg.DistanceMeters
			}
			legs = append(legs, legJSON)
		}
		itineraries = append(itineraries, map[string]any{"id": itinerary.ID, "durationSeconds": itinerary.DurationSeconds, "transfers": itinerary.Transfers, "walkingMeters": itinerary.WalkMeters, "legs": legs})
	}
	return map[string]any{"planId": plan.PlanID, "origin": journeyEndpointJSON(request.Origin), "destination": journeyEndpointJSON(request.Destination), "itineraries": itineraries, "source": plan.Provider, "freshness": map[string]any{"source": plan.Freshness.Source, "fetchedAt": plan.Freshness.FetchedAt.UTC().Format(time.RFC3339), "processedAt": plan.Freshness.ProcessedAt.UTC().Format(time.RFC3339), "isRealtime": plan.Freshness.IsRealtime}}
}
func journeyEndpointJSON(place application.PlaceReference) map[string]any {
	item := map[string]any{"type": string(place.Kind)}
	if place.Coordinate != nil {
		item["coordinate"] = []float64{place.Coordinate.Longitude, place.Coordinate.Latitude}
	}
	if place.Kind == application.PlaceStop {
		item["stopId"] = place.ID
	}
	if place.Label != "" {
		item["label"] = place.Label
	}
	return item
}

// Notification mutations intentionally fail before decoding credentials or
// tokens. This avoids retaining sensitive values in memory or logs while the
// encrypted persistence and push-provider decision gate remain incomplete.
func (s *Server) createInstallation(w http.ResponseWriter, r *http.Request) {
	s.featureUnavailable(w, r, application.FeatureNotifications)
}
func (s *Server) registerPushToken(w http.ResponseWriter, r *http.Request) {
	s.featureUnavailable(w, r, application.FeatureNotifications)
}
func (s *Server) deleteInstallation(w http.ResponseWriter, r *http.Request) {
	s.featureUnavailable(w, r, application.FeatureNotifications)
}
func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	s.featureUnavailable(w, r, application.FeatureNotifications)
}
func (s *Server) createSubscription(w http.ResponseWriter, r *http.Request) {
	s.featureUnavailable(w, r, application.FeatureNotifications)
}
func (s *Server) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	s.featureUnavailable(w, r, application.FeatureNotifications)
}
func (s *Server) nearbyStops(w http.ResponseWriter, r *http.Request) {
	lat, lon, ok := parseCoordinate(r, w)
	if !ok {
		return
	}
	limit, ok := parseLimit(r, w)
	if !ok {
		return
	}
	modes, err := parseModes(r.URL.Query().Get("modes"))
	if err != nil {
		s.validation(w, r, "modes", "invalid")
		return
	}
	radius, ok := boundedInt(r, w, "radiusMeters", 1000, 50, 10000)
	if !ok {
		return
	}
	perMode, ok := boundedInt(r, w, "limitPerMode", min(limit, 20), 1, 20)
	if !ok {
		return
	}
	wheelchair, ok := optionalBool(r, w, "wheelchairAccessible")
	if !ok {
		return
	}
	if _, ok := optionalBool(r, w, "includeArrivals"); !ok {
		return
	}
	items, err := s.app.NearbyStops(r.Context(), application.NearbyQuery{Coordinate: persistence.Coordinate{Longitude: lon, Latitude: lat}, RadiusMeters: radius, LimitPerMode: perMode, Limit: limit, Modes: modes, WheelchairAccessible: wheelchair})
	if err != nil {
		s.staticUnavailable(w, r)
		return
	}
	groups := map[string][]map[string]any{}
	order := []string{}
	for _, stop := range items {
		if _, exists := groups[stop.Mode]; !exists {
			order = append(order, stop.Mode)
		}
		groups[stop.Mode] = append(groups[stop.Mode], map[string]any{"id": stop.ID, "name": stop.Name, "coordinate": []float64{stop.Coordinate.Longitude, stop.Coordinate.Latitude}, "modes": []string{stop.Mode}, "distanceMeters": stop.DistanceMeters})
	}
	out := make([]map[string]any, 0, len(order))
	for _, mode := range order {
		out = append(out, map[string]any{"mode": mode, "stops": groups[mode]})
	}
	s.write(w, http.StatusOK, map[string]any{"distanceType": "straight_line", "groups": out, "freshness": staticFreshness(time.Now().UTC())})
}
func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	id, ok := publicIDParam(r, w, "stop")
	if !ok {
		return
	}
	detail, err := s.app.Stop(r.Context(), id)
	if err != nil {
		s.staticLookupError(w, r, err, "Stop was not found.")
		return
	}
	w.Header().Set("X-Static-Feed-Version", detail.StaticFeedVersion)
	s.etag(w, r, map[string]any{"stop": stopJSON(detail.Stop), "staticFeedVersion": detail.StaticFeedVersion, "freshness": staticFreshness(detail.Freshness.ActivatedAt)})
}
func (s *Server) stopArrivals(w http.ResponseWriter, r *http.Request) {
	id, ok := publicIDParam(r, w, "stop")
	if !ok {
		return
	}
	if _, ok := boundedInt(r, w, "minutes", 90, 1, 240); !ok {
		return
	}
	if v := r.URL.Query().Get("routes"); len(v) > 2000 || hasControl(v) {
		s.validation(w, r, "routes", "invalid")
		return
	}
	if _, ok := optionalDirection(r, w); !ok {
		return
	}
	if _, ok := optionalBool(r, w, "includeScheduled"); !ok {
		return
	}
	routes, err := qualifiedIDs(r.URL.Query().Get("routes"), "route")
	if err != nil {
		s.validation(w, r, "routes", "invalid")
		return
	}
	direction, _ := optionalDirection(r, w)
	minutes, _ := boundedInt(r, w, "minutes", 90, 1, 240)
	include, _ := optionalBool(r, w, "includeScheduled")
	includeScheduled := true
	if include != nil {
		includeScheduled = *include
	}
	arrivals, err := s.app.StopArrivals(r.Context(), persistence.ArrivalFilter{StopID: id, Minutes: minutes, RouteIDs: routes, DirectionID: direction, IncludeScheduled: includeScheduled, Now: time.Now().UTC()})
	if err != nil {
		s.arrivalsUnavailable(w, r)
		return
	}
	items := make([]map[string]any, 0, len(arrivals))
	for _, a := range arrivals {
		items = append(items, arrivalJSON(a))
	}
	s.etag(w, r, map[string]any{"stopId": id, "arrivals": items, "freshness": arrivalsFreshness(arrivals)})
}
func (s *Server) stopSchedule(w http.ResponseWriter, r *http.Request) {
	id, ok := publicIDParam(r, w, "stop")
	if !ok {
		return
	}
	if !validServiceDate(r.URL.Query().Get("serviceDate")) {
		s.validation(w, r, "serviceDate", "required")
		return
	}
	if routeID := r.URL.Query().Get("routeId"); routeID != "" && persistence.ValidatePublicID(routeID, "route") != nil {
		s.validation(w, r, "routeId", "invalid")
		return
	}
	if _, ok := optionalDirection(r, w); !ok {
		return
	}
	q, ok := parseCommon(r, w, true)
	if !ok {
		return
	}
	direction, _ := optionalDirection(r, w)
	schedule, version, next, err := s.app.StopSchedule(r.Context(), persistence.ScheduleFilter{StopID: id, ServiceDate: r.URL.Query().Get("serviceDate"), RouteID: r.URL.Query().Get("routeId"), DirectionID: direction, Limit: q.limit, Cursor: r.URL.Query().Get("cursor")})
	if err != nil {
		s.staticUnavailable(w, r)
		return
	}
	items := make([]map[string]any, 0, len(schedule))
	for _, entry := range schedule {
		items = append(items, scheduleJSON(entry))
	}
	body := map[string]any{"stopId": id, "serviceDate": r.URL.Query().Get("serviceDate"), "staticFeedVersion": version, "schedule": items}
	if next != "" {
		body["nextCursor"] = next
	}
	w.Header().Set("X-Static-Feed-Version", version)
	s.etag(w, r, body)
}
func (s *Server) alerts(w http.ResponseWriter, r *http.Request) {
	q, ok := parseCommon(r, w, true)
	if !ok {
		return
	}
	routes, err := qualifiedIDs(r.URL.Query().Get("routes"), "route")
	if err != nil {
		s.validation(w, r, "routes", "invalid")
		return
	}
	stops, err := qualifiedIDs(r.URL.Query().Get("stops"), "stop")
	if err != nil {
		s.validation(w, r, "stops", "invalid")
		return
	}
	effect := r.URL.Query().Get("effect")
	if effect != "" && !validAlertEffect(effect) {
		s.validation(w, r, "effect", "invalid")
		return
	}
	active := true
	if value, present := r.URL.Query()["active"]; present {
		parsed, err := strconv.ParseBool(value[0])
		if err != nil {
			s.validation(w, r, "active", "invalid")
			return
		}
		active = parsed
	}
	var updated *time.Time
	if raw := r.URL.Query().Get("updatedSince"); raw != "" {
		value, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			s.validation(w, r, "updatedSince", "invalid")
			return
		}
		updated = &value
	}
	alerts, next, err := s.app.Alerts(r.Context(), persistence.AlertFilter{RouteIDs: routes, StopIDs: stops, Modes: q.modes, Effect: effect, Active: active, UpdatedSince: updated, Limit: q.limit, Cursor: r.URL.Query().Get("cursor"), Now: time.Now().UTC()})
	if err != nil {
		s.alertsUnavailable(w, r)
		return
	}
	items := make([]map[string]any, 0, len(alerts))
	for _, alert := range alerts {
		items = append(items, alertJSON(alert))
	}
	body := map[string]any{"alerts": items, "freshness": alertsFreshness(alerts)}
	if next != "" {
		body["nextCursor"] = next
	}
	s.etag(w, r, body)
}
func (s *Server) alert(w http.ResponseWriter, r *http.Request) {
	id, ok := publicIDParam(r, w, "alert")
	if !ok {
		return
	}
	item, err := s.app.Alert(r.Context(), id)
	if errors.Is(err, persistence.ErrNotFound) {
		s.error(w, r, http.StatusNotFound, "not_found", "Alert was not found.", nil)
		return
	}
	if err != nil {
		s.alertsUnavailable(w, r)
		return
	}
	s.etag(w, r, map[string]any{"alert": alertJSON(item)})
}
func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	id, ok := publicIDParam(r, w, "route")
	if !ok {
		return
	}
	if date := r.URL.Query().Get("serviceDate"); date != "" && !validServiceDate(date) {
		s.validation(w, r, "serviceDate", "invalid")
		return
	}
	detail, err := s.app.Route(r.Context(), id)
	if err != nil {
		s.staticLookupError(w, r, err, "Route was not found.")
		return
	}
	directions := make([]map[string]any, 0, len(detail.Directions))
	for _, d := range detail.Directions {
		x := map[string]any{"directionId": d.ID}
		if d.Headsign != "" {
			x["headsign"] = d.Headsign
		}
		directions = append(directions, x)
	}
	w.Header().Set("X-Static-Feed-Version", detail.StaticFeedVersion)
	s.etag(w, r, map[string]any{"route": routeJSON(detail.Route), "directions": directions, "staticFeedVersion": detail.StaticFeedVersion, "freshness": staticFreshness(detail.Freshness.ActivatedAt)})
}
func (s *Server) routeShape(w http.ResponseWriter, r *http.Request) {
	id, ok := publicIDParam(r, w, "route")
	if !ok {
		return
	}
	direction, ok := optionalDirection(r, w)
	if !ok {
		return
	}
	if _, err := s.app.Route(r.Context(), id); err != nil {
		s.staticLookupError(w, r, err, "Route was not found.")
		return
	}
	shapes, version, err := s.app.RouteShapes(r.Context(), id, direction)
	if err != nil {
		s.staticUnavailable(w, r)
		return
	}
	features := make([]any, 0, len(shapes))
	for _, shape := range shapes {
		properties := map[string]any{"shapeId": shape.ID, "routeId": shape.RouteID}
		if shape.DirectionID != nil {
			properties["directionId"] = *shape.DirectionID
		}
		features = append(features, map[string]any{"type": "Feature", "geometry": map[string]any{"type": "LineString", "coordinates": shape.Coordinates}, "properties": properties})
	}
	w.Header().Set("Content-Type", "application/geo+json; charset=utf-8")
	w.Header().Set("X-Static-Feed-Version", version)
	s.etag(w, r, map[string]any{"type": "FeatureCollection", "features": features, "staticFeedVersion": version})
}
func (s *Server) routeStops(w http.ResponseWriter, r *http.Request) {
	id, ok := publicIDParam(r, w, "route")
	if !ok {
		return
	}
	if date := r.URL.Query().Get("serviceDate"); date != "" && !validServiceDate(date) {
		s.validation(w, r, "serviceDate", "invalid")
		return
	}
	direction, ok := optionalDirection(r, w)
	if !ok {
		return
	}
	if _, err := s.app.Route(r.Context(), id); err != nil {
		s.staticLookupError(w, r, err, "Route was not found.")
		return
	}
	stops, version, err := s.app.RouteStops(r.Context(), id, direction)
	if err != nil {
		s.staticUnavailable(w, r)
		return
	}
	out := make([]any, 0, len(stops))
	for _, stop := range stops {
		x := stopJSON(stop.Stop)
		x["sequence"] = stop.Sequence
		out = append(out, x)
	}
	body := map[string]any{"routeId": id, "stops": out, "staticFeedVersion": version}
	if direction != nil {
		body["directionId"] = *direction
	}
	w.Header().Set("X-Static-Feed-Version", version)
	s.etag(w, r, body)
}
func (s *Server) routeVehicles(w http.ResponseWriter, r *http.Request) {
	id, ok := publicIDParam(r, w, "route")
	if !ok {
		return
	}
	if _, ok := optionalDirection(r, w); !ok {
		return
	}
	if _, err := s.app.Route(r.Context(), id); err != nil {
		s.staticLookupError(w, r, err, "Route was not found.")
		return
	}
	vs, err := s.app.ListVehicles(r.Context(), nil)
	if err != nil {
		s.unavailable(w, r)
		return
	}
	filtered := make([]persistence.Vehicle, 0, len(vs))
	for _, v := range vs {
		if v.RouteID != nil && *v.RouteID == id {
			filtered = append(filtered, v)
		}
	}
	s.etag(w, r, vehicleCollection(filtered, time.Now().UTC()))
}

type requestIDKey struct{}

func requestID(ctx context.Context) string { v, _ := ctx.Value(requestIDKey{}).(string); return v }
func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if !validRequestID(id) {
			id = newRequestID()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

// Request IDs are safe, bounded correlation values. We only propagate a
// caller-supplied value when it is a simple token so it remains safe for HTTP
// headers and structured logs; arbitrary user input must never become a log
// correlation field.
func validRequestID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !safeTokenRune(r, "._-") {
			return false
		}
	}
	return true
}

func newRequestID() string {
	var token [16]byte
	if _, err := cryptorand.Read(token[:]); err == nil {
		return "req_" + hex.EncodeToString(token[:])
	}
	// crypto/rand failures are exceptionally rare. Preserve request handling
	// without reflecting untrusted input; the nanosecond suffix keeps this
	// fallback distinguishable for local diagnostics.
	return fmt.Sprintf("req_fallback_%d", time.Now().UnixNano())
}
func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				s.error(w, r, http.StatusInternalServerError, "internal_error", "An unexpected error occurred.", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// cors permits only explicitly configured local-development origins. Browser
// production traffic is same-origin through Caddy; wildcard origins and
// credentials are intentionally never emitted.
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimRight(r.Header.Get("Origin"), "/")
		allowed := false
		for _, candidate := range s.config.API.AllowedWebOrigins {
			if origin == candidate {
				allowed = true
				break
			}
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, If-None-Match")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.Header().Add("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) live(w http.ResponseWriter, r *http.Request) {
	s.write(w, http.StatusOK, map[string]string{"status": "ok", "requestId": requestID(r.Context())})
}
func (s *Server) readiness(w http.ResponseWriter, r *http.Request) {
	if s.ready != nil {
		if err := s.ready(r.Context()); err != nil {
			s.error(w, r, http.StatusServiceUnavailable, "temporarily_unavailable", "Core data is not available.", nil)
			return
		}
	} else if s.app.Vehicles == nil {
		s.error(w, r, http.StatusServiceUnavailable, "temporarily_unavailable", "Core data is not available.", nil)
		return
	}
	s.live(w, r)
}
func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	c := s.config.API
	planner := map[string]any{"enabled": s.app.Planning.Planner != nil}
	if s.app.Planning.Planner == nil {
		planner["reason"] = "planner_not_configured"
	}
	body := map[string]any{"apiVersion": c.Version, "minimumAppVersion": c.MinimumAppVersion, "features": map[string]any{"vehicleMap": map[string]any{"enabled": true}, "placeSearch": map[string]any{"enabled": false, "reason": "external_provider_gate_pending"}, "journeyPlanner": planner, "notifications": map[string]any{"enabled": false, "reason": application.ReasonNotificationGatePending}}, "sources": map[string]any{"trimetGtfsRt": map[string]any{"enabled": true}, "trimetStreetcar": map[string]any{"enabled": true}}, "pollingRecommendations": map[string]int{"vehiclesSeconds": 15}, "staleThresholdSeconds": map[string]int{"vehicles": c.StaleThresholdSeconds}, "serviceBounds": map[string]any{"bbox": []float64{-123, 45.3, -122.3, 45.8}}, "staticFeed": map[string]any{"version": c.StaticFeedVersion, "publishedAt": c.StaticFeedPublishedAt.UTC().Format(time.RFC3339)}}
	w.Header().Set("X-Api-Version", c.Version)
	s.etag(w, r, body)
}
func (s *Server) routes(w http.ResponseWriter, r *http.Request) {
	q, ok := parseCommon(r, w, true)
	if !ok {
		return
	}
	query, serviceDate := r.URL.Query().Get("query"), r.URL.Query().Get("serviceDate")
	if (query != "" && (len(query) > 100 || hasControl(query))) || (serviceDate != "" && !validServiceDate(serviceDate)) {
		s.validation(w, r, "query", "invalid")
		return
	}
	page, err := s.app.ListRoutes(r.Context(), application.RouteQuery{Modes: q.modes, Query: query, ServiceDate: serviceDate, Limit: q.limit, Cursor: r.URL.Query().Get("cursor")})
	if err != nil {
		s.unavailable(w, r)
		return
	}
	out := make([]map[string]any, 0, len(page.Items))
	for _, x := range page.Items {
		m := map[string]any{"id": x.ID, "mode": x.Mode, "shortName": x.ShortName, "longName": x.LongName}
		if x.Color != nil {
			m["color"] = *x.Color
		}
		if x.TextColor != nil {
			m["textColor"] = *x.TextColor
		}
		out = append(out, m)
	}
	body := map[string]any{"routes": out, "staticFeedVersion": page.StaticFeedVersion}
	if page.NextCursor != "" {
		body["nextCursor"] = page.NextCursor
	}
	w.Header().Set("X-Static-Feed-Version", page.StaticFeedVersion)
	s.etag(w, r, body)
}
func (s *Server) stops(w http.ResponseWriter, r *http.Request) {
	q, ok := parseCommon(r, w, true)
	if !ok {
		return
	}
	if v := r.URL.Query().Get("query"); v == "" || len(v) > 100 || hasControl(v) {
		s.validation(w, r, "query", "required")
		return
	}
	page, err := s.app.ListStops(r.Context(), application.StopQuery{Modes: q.modes, Query: r.URL.Query().Get("query"), Limit: q.limit, Cursor: r.URL.Query().Get("cursor")})
	if err != nil {
		s.unavailable(w, r)
		return
	}
	out := make([]map[string]any, 0, len(page.Items))
	for _, x := range page.Items {
		m := map[string]any{"id": x.ID, "name": x.Name, "coordinate": []float64{x.Coordinate.Longitude, x.Coordinate.Latitude}, "modes": x.Modes}
		if len(x.RouteIDs) > 0 {
			m["routeIds"] = x.RouteIDs
		}
		if x.ParentStopID != nil {
			m["parentStopId"] = *x.ParentStopID
		}
		if x.WheelchairAccessible != nil {
			m["wheelchairAccessible"] = *x.WheelchairAccessible
		}
		out = append(out, m)
	}
	body := map[string]any{"stops": out}
	if page.NextCursor != "" {
		body["nextCursor"] = page.NextCursor
	}
	s.write(w, http.StatusOK, body)
}
func (s *Server) vehicles(w http.ResponseWriter, r *http.Request) {
	v, ok := s.vehicleFilter(r, w)
	if !ok {
		return
	}
	vs, err := s.app.ListVehicles(r.Context(), v.sources)
	if err != nil {
		s.unavailable(w, r)
		return
	}
	vs = filterVehicles(vs, v)
	if len(vs) > 0 {
		w.Header().Set("X-Data-Freshness", string(vs[0].Freshness))
	}
	body := vehicleCollection(vs, time.Now().UTC())
	if v.geojson {
		body = vehicleGeoJSON(vs, time.Now().UTC())
	}
	s.etag(w, r, body)
}
func (s *Server) vehicleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" || len(q) > 100 || hasControl(q) || !validSearch(q) {
		s.validation(w, r, "q", "invalid")
		return
	}
	limit, ok := parseLimit(r, w)
	if !ok {
		return
	}
	vs, err := s.app.SearchVehicles(r.Context(), q, limit)
	if err != nil {
		s.unavailable(w, r)
		return
	}
	s.write(w, http.StatusOK, map[string]any{"query": q, "vehicles": vehiclesJSON(vs, time.Now().UTC()), "freshness": collectionFreshness(vs, time.Now().UTC())})
}
func (s *Server) vehicle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := persistence.ValidatePublicID(id, "vehicle"); err != nil {
		s.validation(w, r, "id", "invalid")
		return
	}
	v, err := s.app.Vehicle(r.Context(), id)
	if errors.Is(err, application.ErrUnavailable) {
		s.unavailable(w, r)
		return
	}
	if err != nil {
		s.error(w, r, http.StatusNotFound, "not_found", "Vehicle was not found.", nil)
		return
	}
	s.etag(w, r, map[string]any{"vehicle": vehicleJSON(v, time.Now().UTC())})
}
func (s *Server) vehicleHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := persistence.ValidatePublicID(id, "vehicle"); err != nil {
		s.validation(w, r, "id", "invalid")
		return
	}
	query, ok := s.parseVehicleHistoryQuery(r, w)
	if !ok {
		return
	}
	query.VehicleID = id
	items, err := s.app.VehicleHistory(r.Context(), query)
	if errors.Is(err, application.ErrUnavailable) {
		s.unavailable(w, r)
		return
	}
	if err != nil {
		s.unavailable(w, r)
		return
	}
	observations := make([]map[string]any, 0, len(items))
	for _, item := range items {
		entry := map[string]any{"coordinate": []float64{item.Coordinate.Longitude, item.Coordinate.Latitude}, "observedAt": item.ObservedAt.UTC().Format(time.RFC3339), "mode": item.Mode, "freshness": map[string]any{"status": item.Freshness, "fetchedAt": item.FetchedAt.UTC().Format(time.RFC3339)}}
		if item.RouteID != nil {
			entry["routeId"] = *item.RouteID
		}
		if item.TripID != nil {
			entry["tripId"] = *item.TripID
		}
		observations = append(observations, entry)
	}
	body := map[string]any{"vehicleId": id, "observations": observations, "retentionDays": 30, "freshness": map[string]any{"status": "historical", "source": "normalized-vehicle-observations"}}
	if len(items) == query.Limit {
		body["nextCursor"] = items[len(items)-1].ObservedAt.UTC().Format(time.RFC3339Nano)
	}
	s.etag(w, r, body)
}

func (s *Server) parseVehicleHistoryQuery(r *http.Request, w http.ResponseWriter) (persistence.VehicleHistoryFilter, bool) {
	now := time.Now().UTC()
	to := now
	from := now.Add(-24 * time.Hour)
	if raw := r.URL.Query().Get("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			s.validation(w, r, "to", "invalid")
			return persistence.VehicleHistoryFilter{}, false
		}
		to = parsed.UTC()
	}
	if raw := r.URL.Query().Get("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			s.validation(w, r, "from", "invalid")
			return persistence.VehicleHistoryFilter{}, false
		}
		from = parsed.UTC()
	}
	if from.After(to) || to.Sub(from) > 30*24*time.Hour || from.Before(now.Add(-30*24*time.Hour)) {
		s.validation(w, r, "window", "invalid")
		return persistence.VehicleHistoryFilter{}, false
	}
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 500 {
			s.validation(w, r, "limit", "invalid")
			return persistence.VehicleHistoryFilter{}, false
		}
		limit = parsed
	}
	var cursor *time.Time
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil || parsed.Before(from) || parsed.After(to) {
			s.validation(w, r, "cursor", "invalid")
			return persistence.VehicleHistoryFilter{}, false
		}
		utc := parsed.UTC()
		cursor = &utc
	}
	return persistence.VehicleHistoryFilter{From: from, To: to, Limit: limit, Cursor: cursor}, true
}

func (s *Server) unavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Retry-After", "30")
	s.error(w, r, http.StatusServiceUnavailable, "source_unavailable", "Vehicle positions are temporarily unavailable.", map[string]any{"retryAfterSeconds": 30, "source": "normalized-vehicle-positions"})
}
func (s *Server) featureUnavailable(w http.ResponseWriter, r *http.Request, feature string) {
	w.Header().Set("Retry-After", "3600")
	s.error(w, r, http.StatusServiceUnavailable, "feature_unavailable", "This optional feature is disabled pending its documented decision gate.", map[string]any{"retryAfterSeconds": 3600, "source": "feature-gate:" + feature})
}
func (s *Server) etag(w http.ResponseWriter, r *http.Request, body any) {
	b, _ := json.Marshal(body)
	sum := sha256.Sum256(b)
	tag := "\"" + hex.EncodeToString(sum[:]) + "\""
	w.Header().Set("ETag", tag)
	if r.Header.Get("If-None-Match") == tag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	_, _ = w.Write(b)
}
func parseCoordinate(r *http.Request, w http.ResponseWriter) (float64, float64, bool) {
	latitude := r.URL.Query().Get("latitude")
	longitude := r.URL.Query().Get("longitude")
	// Accept the documented compact spellings too while existing clients move
	// to the canonical latitude/longitude names.
	if latitude == "" {
		latitude = r.URL.Query().Get("lat")
	}
	if longitude == "" {
		longitude = r.URL.Query().Get("lon")
	}
	lat, e1 := strconv.ParseFloat(latitude, 64)
	lon, e2 := strconv.ParseFloat(longitude, 64)
	if e1 != nil || e2 != nil || math.IsNaN(lat) || math.IsNaN(lon) || math.IsInf(lat, 0) || math.IsInf(lon, 0) || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		writeValidation(w, r, "latitude", "invalid")
		return 0, 0, false
	}
	return lat, lon, true
}
func boundedInt(r *http.Request, w http.ResponseWriter, name string, fallback, min, max int) (int, bool) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return fallback, true
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < min || n > max {
		(&Server{}).validation(w, r, name, "invalid")
		return 0, false
	}
	return n, true
}
func optionalBool(r *http.Request, w http.ResponseWriter, name string) (*bool, bool) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return nil, true
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		(&Server{}).validation(w, r, name, "invalid")
		return nil, false
	}
	return &parsed, true
}
func optionalDirection(r *http.Request, w http.ResponseWriter) (*int, bool) {
	v := r.URL.Query().Get("directionId")
	if v == "" {
		return nil, true
	}
	n, err := strconv.Atoi(v)
	if err != nil || (n != 0 && n != 1) {
		(&Server{}).validation(w, r, "directionId", "invalid")
		return nil, false
	}
	return &n, true
}
func publicIDParam(r *http.Request, w http.ResponseWriter, kind string) (string, bool) {
	id := chi.URLParam(r, "id")
	if persistence.ValidatePublicID(id, kind) != nil {
		(&Server{}).validation(w, r, "id", "invalid")
		return "", false
	}
	return id, true
}
func routeJSON(x persistence.CatalogRoute) map[string]any {
	out := map[string]any{"id": x.ID, "mode": x.Mode, "shortName": x.ShortName, "longName": x.LongName}
	if x.Color != nil {
		out["color"] = *x.Color
	}
	if x.TextColor != nil {
		out["textColor"] = *x.TextColor
	}
	return out
}
func stopJSON(x persistence.CatalogStop) map[string]any {
	out := map[string]any{"id": x.ID, "name": x.Name, "coordinate": []float64{x.Coordinate.Longitude, x.Coordinate.Latitude}, "modes": x.Modes}
	if len(x.RouteIDs) > 0 {
		out["routeIds"] = x.RouteIDs
	}
	if x.ParentStopID != nil {
		out["parentStopId"] = *x.ParentStopID
	}
	if x.WheelchairAccessible != nil {
		out["wheelchairAccessible"] = *x.WheelchairAccessible
	}
	return out
}
func staticFreshness(activated time.Time) map[string]any {
	now := time.Now().UTC()
	if activated.IsZero() {
		activated = now
	}
	return map[string]any{"source": "normalized-static-gtfs", "fetchedAt": activated.UTC().Format(time.RFC3339), "processedAt": activated.UTC().Format(time.RFC3339), "status": "fresh", "ageSeconds": int(maxDuration(0, now.Sub(activated)).Seconds()), "isRealtime": false}
}
func (s *Server) staticUnavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Retry-After", "30")
	s.error(w, r, http.StatusServiceUnavailable, "source_unavailable", "Normalized static transit data is temporarily unavailable.", map[string]any{"retryAfterSeconds": 30, "source": "normalized-static-gtfs"})
}
func (s *Server) arrivalsUnavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Retry-After", "30")
	s.error(w, r, http.StatusServiceUnavailable, "source_unavailable", "Normalized arrivals are temporarily unavailable.", map[string]any{"retryAfterSeconds": 30, "source": "normalized-arrivals"})
}
func (s *Server) alertsUnavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Retry-After", "30")
	s.error(w, r, http.StatusServiceUnavailable, "source_unavailable", "Normalized alerts are temporarily unavailable.", map[string]any{"retryAfterSeconds": 30, "source": "normalized-alerts"})
}
func qualifiedIDs(raw, kind string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	values := strings.Split(raw, ",")
	for _, value := range values {
		if persistence.ValidatePublicID(value, kind) != nil {
			return nil, persistence.ErrInvalidPublicID
		}
	}
	return values, nil
}
func validAlertEffect(value string) bool {
	for _, allowed := range []string{"no_service", "reduced_service", "significant_delays", "detour", "stop_moved", "modified_service", "other", "unknown"} {
		if value == allowed {
			return true
		}
	}
	return false
}
func scheduleJSON(x persistence.ScheduleTime) map[string]any {
	out := map[string]any{"tripId": x.TripID, "routeId": x.RouteID, "stopId": x.StopID, "serviceDate": x.ServiceDate, "serviceDaySeconds": x.ServiceDaySeconds}
	if !x.DepartureAt.IsZero() {
		out["departureAt"] = x.DepartureAt.UTC().Format(time.RFC3339)
	}
	if x.DirectionID != nil {
		out["directionId"] = *x.DirectionID
	}
	if x.Headsign != "" {
		out["headsign"] = x.Headsign
	}
	return out
}
func arrivalJSON(x persistence.Arrival) map[string]any {
	out := map[string]any{"id": x.ID, "stopId": x.StopID, "routeId": x.RouteID, "status": x.Status, "scheduledAt": x.ScheduledAt.UTC().Format(time.RFC3339), "freshness": staticFreshness(x.Freshness.ActivatedAt)}
	if x.DirectionID != nil {
		out["directionId"] = *x.DirectionID
	}
	if x.Headsign != "" {
		out["headsign"] = x.Headsign
	}
	if x.EstimatedAt != nil {
		out["estimatedAt"] = x.EstimatedAt.UTC().Format(time.RFC3339)
	}
	if x.TripID != nil {
		out["tripId"] = *x.TripID
	}
	if x.StopSequence != nil {
		out["stopSequence"] = *x.StopSequence
	}
	return out
}
func alertJSON(x persistence.Alert) map[string]any {
	periods := []map[string]any{{}}
	if x.ActiveFrom != nil {
		periods[0]["startAt"] = x.ActiveFrom.UTC().Format(time.RFC3339)
	}
	if x.ActiveUntil != nil {
		periods[0]["endAt"] = x.ActiveUntil.UTC().Format(time.RFC3339)
	}
	out := map[string]any{"id": x.ID, "revision": x.Revision, "header": x.Header, "periods": periods, "source": x.Source, "freshness": map[string]any{"source": x.Source, "fetchedAt": x.FetchedAt.UTC().Format(time.RFC3339), "processedAt": x.ProcessedAt.UTC().Format(time.RFC3339), "status": "unknown", "ageSeconds": int(maxDuration(0, time.Since(x.FetchedAt)).Seconds()), "isRealtime": true}}
	if x.Description != "" {
		out["description"] = x.Description
	}
	if x.Cause != "" {
		out["cause"] = x.Cause
	}
	if validAlertEffect(x.Effect) {
		out["effect"] = x.Effect
	}
	if x.SourceURL != "" {
		out["sourceUrl"] = x.SourceURL
	}
	return out
}
func arrivalsFreshness(items []persistence.Arrival) map[string]any {
	if len(items) == 0 {
		return map[string]any{"source": "normalized-arrivals", "fetchedAt": time.Now().UTC().Format(time.RFC3339), "processedAt": time.Now().UTC().Format(time.RFC3339), "status": "unknown", "ageSeconds": 0, "isRealtime": false}
	}
	return staticFreshness(items[0].Freshness.ActivatedAt)
}
func alertsFreshness(items []persistence.Alert) map[string]any {
	if len(items) == 0 {
		return map[string]any{"source": "normalized-alerts", "fetchedAt": time.Now().UTC().Format(time.RFC3339), "processedAt": time.Now().UTC().Format(time.RFC3339), "status": "unknown", "ageSeconds": 0, "isRealtime": true}
	}
	return alertJSON(items[0])["freshness"].(map[string]any)
}
func (s *Server) staticLookupError(w http.ResponseWriter, r *http.Request, err error, message string) {
	if errors.Is(err, persistence.ErrNotFound) {
		s.error(w, r, http.StatusNotFound, "not_found", message, nil)
		return
	}
	s.staticUnavailable(w, r)
}
func (s *Server) write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
func (s *Server) error(w http.ResponseWriter, r *http.Request, status int, code, msg string, extra map[string]any) {
	e := map[string]any{"code": code, "message": msg, "requestId": requestID(r.Context())}
	for k, v := range extra {
		e[k] = v
	}
	s.write(w, status, map[string]any{"error": e})
}

type common struct {
	modes []string
	limit int
}

func parseCommon(r *http.Request, w http.ResponseWriter, allowCursor bool) (common, bool) {
	limit, ok := parseLimit(r, w)
	if !ok {
		return common{}, false
	}
	modes, err := parseModes(r.URL.Query().Get("modes"))
	if err != nil {
		writeValidation(w, r, "modes", "invalid")
		return common{}, false
	}
	if c := r.URL.Query().Get("cursor"); (!allowCursor && c != "") || len(c) > 1000 || hasControl(c) {
		writeValidation(w, r, "cursor", "invalid")
		return common{}, false
	}
	return common{modes, limit}, true
}
func parseLimit(r *http.Request, w http.ResponseWriter) (int, bool) {
	v := r.URL.Query().Get("limit")
	if v == "" {
		return 50, true
	}
	n, e := strconv.Atoi(v)
	if e != nil || n < 1 || n > 100 {
		writeValidation(w, r, "limit", "invalid")
		return 0, false
	}
	return n, true
}
func parseModes(v string) ([]string, error) {
	if v == "" {
		return nil, nil
	}
	allowed := map[string]bool{"bus": true, "light_rail": true, "streetcar": true, "commuter_rail": true, "ferry": true, "unknown": true}
	out := strings.Split(v, ",")
	for _, x := range out {
		if !allowed[x] {
			return nil, errors.New("bad mode")
		}
	}
	return out, nil
}
func hasControl(v string) bool { return strings.ContainsAny(v, "\r\n\t") }
func safeTokenRune(r rune, extra string) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune(extra, r)
}
func validSearch(v string) bool {
	for _, r := range v {
		if !safeTokenRune(r, ":_-") {
			return false
		}
	}
	return true
}
func validServiceDate(v string) bool { _, err := time.Parse("2006-01-02", v); return err == nil }
func writeValidation(w http.ResponseWriter, r *http.Request, field, detail string) {
	(&Server{}).error(w, r, http.StatusBadRequest, "validation_error", "Request validation failed.", map[string]any{"details": []map[string]string{{"field": field, "code": detail}}})
}
func (s *Server) validation(w http.ResponseWriter, r *http.Request, field, detail string) {
	s.error(w, r, http.StatusBadRequest, "validation_error", "Request validation failed.", map[string]any{"details": []map[string]string{{"field": field, "code": detail}}})
}

type vehicleQuery struct {
	modes, sources, routes []string
	freshness              string
	geojson                bool
	bounds                 *boundingBox
}

// boundingBox uses the API's WGS84 west,south,east,north ordering. It is
// deliberately parsed at the HTTP boundary: callers never pass arbitrary
// coordinate strings into persistence or source adapters.
type boundingBox struct{ west, south, east, north float64 }

func (s *Server) vehicleFilter(r *http.Request, w http.ResponseWriter) (vehicleQuery, bool) {
	modes, e := parseModes(r.URL.Query().Get("modes"))
	if e != nil {
		s.validation(w, r, "modes", "invalid")
		return vehicleQuery{}, false
	}
	format := r.URL.Query().Get("format")
	if format != "" && format != "json" && format != "geojson" {
		s.validation(w, r, "format", "invalid")
		return vehicleQuery{}, false
	}
	fq := r.URL.Query().Get("freshness")
	if fq != "" && fq != "fresh" && fq != "aging" && fq != "stale" && fq != "unknown" {
		s.validation(w, r, "freshness", "invalid")
		return vehicleQuery{}, false
	}
	for _, name := range []string{"routes", "sources", "directions", "bbox"} {
		if len(r.URL.Query().Get(name)) > 2000 || hasControl(r.URL.Query().Get(name)) {
			s.validation(w, r, name, "invalid")
			return vehicleQuery{}, false
		}
	}
	bounds, err := parseBoundingBox(r.URL.Query().Get("bbox"))
	if err != nil {
		s.validation(w, r, "bbox", "invalid")
		return vehicleQuery{}, false
	}
	return vehicleQuery{modes: modes, sources: split(r.URL.Query().Get("sources")), routes: split(r.URL.Query().Get("routes")), freshness: fq, geojson: format == "geojson", bounds: bounds}, true
}

func parseBoundingBox(raw string) (*boundingBox, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return nil, errors.New("bbox must have four coordinates")
	}
	values := [4]float64{}
	for i, part := range parts {
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	bounds := &boundingBox{west: values[0], south: values[1], east: values[2], north: values[3]}
	if bounds.west < -180 || bounds.west > 180 || bounds.east < -180 || bounds.east > 180 || bounds.south < -90 || bounds.south > 90 || bounds.north < -90 || bounds.north > 90 || bounds.west > bounds.east || bounds.south > bounds.north {
		return nil, errors.New("invalid WGS84 bounds")
	}
	return bounds, nil
}

func (b boundingBox) contains(coordinate persistence.Coordinate) bool {
	return coordinate.Longitude >= b.west && coordinate.Longitude <= b.east && coordinate.Latitude >= b.south && coordinate.Latitude <= b.north
}
func split(v string) []string {
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}
func filterVehicles(vs []persistence.Vehicle, q vehicleQuery) []persistence.Vehicle {
	out := vs[:0]
	for _, v := range vs {
		if len(q.modes) > 0 && !contains(q.modes, v.Mode) {
			continue
		}
		if len(q.routes) > 0 && (v.RouteID == nil || !contains(q.routes, *v.RouteID)) {
			continue
		}
		if q.freshness != "" && q.freshness != string(v.Freshness) {
			continue
		}
		if q.bounds != nil && !q.bounds.contains(v.Coordinate) {
			continue
		}
		out = append(out, v)
	}
	return out
}
func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func vehiclesJSON(vs []persistence.Vehicle, now time.Time) []any {
	o := make([]any, len(vs))
	for i, v := range vs {
		o[i] = vehicleJSON(v, now)
	}
	return o
}
func vehicleJSON(v persistence.Vehicle, now time.Time) map[string]any {
	m := map[string]any{"id": v.ID, "sourceVehicleId": v.SourceVehicleID, "mode": v.Mode, "coordinate": []float64{v.Coordinate.Longitude, v.Coordinate.Latitude}, "inService": true, "freshness": freshness(v, now)}
	if v.RouteID != nil {
		m["routeId"] = *v.RouteID
	}
	if v.TripID != nil {
		m["tripId"] = *v.TripID
	}
	return m
}
func freshness(v persistence.Vehicle, now time.Time) map[string]any {
	age := maxDuration(0, now.Sub(v.FetchedAt))
	status := v.Freshness
	// Stored ingest status is only a snapshot. Re-evaluate it at read time so
	// an outage cannot leave a previously fresh projection reported as fresh.
	if age > 90*time.Second {
		status = persistence.FreshnessStale
	} else if age > 45*time.Second && status == persistence.FreshnessFresh {
		status = persistence.FreshnessAging
	}
	m := map[string]any{"source": v.SourceID, "fetchedAt": v.FetchedAt.UTC().Format(time.RFC3339), "processedAt": v.ProcessedAt.UTC().Format(time.RFC3339), "status": status, "ageSeconds": int(age.Seconds()), "isRealtime": true}
	if v.SourceUpdatedAt != nil {
		m["sourceUpdatedAt"] = v.SourceUpdatedAt.UTC().Format(time.RFC3339)
	}
	if v.EntityUpdatedAt != nil {
		m["entityUpdatedAt"] = v.EntityUpdatedAt.UTC().Format(time.RFC3339)
	}
	return m
}
func collectionFreshness(vs []persistence.Vehicle, now time.Time) map[string]any {
	if len(vs) == 0 {
		return map[string]any{"source": "normalized-vehicle-positions", "fetchedAt": now.Format(time.RFC3339), "processedAt": now.Format(time.RFC3339), "status": "unknown", "ageSeconds": 0, "isRealtime": true}
	}
	return freshness(vs[0], now)
}
func vehicleCollection(vs []persistence.Vehicle, now time.Time) map[string]any {
	snap := ""
	if len(vs) > 0 {
		snap = fmt.Sprintf("%d", vs[0].SnapshotID)
	}
	return map[string]any{"snapshotId": snap, "vehicles": vehiclesJSON(vs, now), "freshness": collectionFreshness(vs, now)}
}
func vehicleGeoJSON(vs []persistence.Vehicle, now time.Time) map[string]any {
	fs := make([]any, len(vs))
	for i, v := range vs {
		fs[i] = map[string]any{"type": "Feature", "geometry": map[string]any{"type": "Point", "coordinates": []float64{v.Coordinate.Longitude, v.Coordinate.Latitude}}, "properties": vehicleJSON(v, now)}
	}
	snap := ""
	if len(vs) > 0 {
		snap = fmt.Sprintf("%d", vs[0].SnapshotID)
	}
	return map[string]any{"type": "FeatureCollection", "features": fs, "snapshotId": snap, "freshness": collectionFreshness(vs, now)}
}
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

type rateLimiter struct {
	mu      sync.Mutex
	cfg     config.RateLimit
	buckets map[string]*bucket
}
type bucket struct {
	n     int
	start time.Time
}

func newRateLimiter(c config.RateLimit) *rateLimiter {
	return &rateLimiter{cfg: c, buckets: map[string]*bucket{}}
}
func (s *Server) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := clientAddress(r)
		if !s.limiter.allow(host) {
			w.Header().Set("Retry-After", strconv.Itoa(int(s.limiter.cfg.Window.Seconds())))
			s.error(w, r, http.StatusTooManyRequests, "rate_limited", "Too many requests.", map[string]any{"retryAfterSeconds": int(s.limiter.cfg.Window.Seconds())})
			return
		}
		next.ServeHTTP(w, r)
	})
}
func clientAddress(r *http.Request) string {
	// The API is reachable only through Caddy in production; use the first
	// forwarded address that Caddy supplies so all riders do not share its
	// container address. Direct development requests retain RemoteAddr.
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); net.ParseIP(forwarded) != nil {
		return forwarded
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	// Expire inactive keys while servicing traffic, bounding memory even when
	// an attacker rotates source addresses.
	for existing, entry := range l.buckets {
		if now.Sub(entry.start) >= l.cfg.Window {
			delete(l.buckets, existing)
		}
	}
	b := l.buckets[key]
	if b == nil || now.Sub(b.start) >= l.cfg.Window {
		l.buckets[key] = &bucket{n: 1, start: now}
		return true
	}
	if b.n >= l.cfg.Requests {
		return false
	}
	b.n++
	return true
}
