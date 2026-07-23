// Package api implements the public HTTP boundary. It never logs request query
// values or coordinates, because both can be sensitive rider data.
package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	r.Use(s.requestID, s.recover, s.securityHeaders, s.limit)
	r.Get("/health/live", s.live)
	r.Get("/health/ready", s.readiness)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/config", s.getConfig)
		r.Get("/routes", s.routes)
		r.Get("/stops", s.stops)
		r.Get("/vehicles", s.vehicles)
		r.Get("/vehicles/search", s.vehicleSearch)
		r.Get("/vehicles/{id}", s.vehicle)
	})
	return r
}

type requestIDKey struct{}

func requestID(ctx context.Context) string { v, _ := ctx.Value(requestIDKey{}).(string); return v }
func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if len(id) == 0 || len(id) > 128 || strings.ContainsAny(id, "\r\n") {
			id = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
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
	body := map[string]any{"apiVersion": c.Version, "minimumAppVersion": c.MinimumAppVersion, "features": map[string]any{"vehicleMap": map[string]any{"enabled": true}}, "sources": map[string]any{"trimetGtfsRt": map[string]any{"enabled": true}}, "pollingRecommendations": map[string]int{"vehiclesSeconds": 15}, "staleThresholdSeconds": map[string]int{"vehicles": c.StaleThresholdSeconds}, "serviceBounds": map[string]any{"bbox": []float64{-123, 45.3, -122.3, 45.8}}, "staticFeed": map[string]any{"version": c.StaticFeedVersion, "publishedAt": c.StaticFeedPublishedAt.UTC().Format(time.RFC3339)}}
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
func (s *Server) unavailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Retry-After", "30")
	s.error(w, r, http.StatusServiceUnavailable, "source_unavailable", "Vehicle positions are temporarily unavailable.", map[string]any{"retryAfterSeconds": 30, "source": "normalized-vehicle-positions"})
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
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(b)
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
func validSearch(v string) bool {
	for _, r := range v {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune(":_-", r)) {
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
}

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
	return vehicleQuery{modes: modes, sources: split(r.URL.Query().Get("sources")), routes: split(r.URL.Query().Get("routes")), freshness: fq, geojson: format == "geojson"}, true
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
	m := map[string]any{"source": v.SourceID, "fetchedAt": v.FetchedAt.UTC().Format(time.RFC3339), "processedAt": v.ProcessedAt.UTC().Format(time.RFC3339), "status": v.Freshness, "ageSeconds": int(maxDuration(0, now.Sub(v.FetchedAt)).Seconds()), "isRealtime": true}
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
		host, _, e := net.SplitHostPort(r.RemoteAddr)
		if e != nil {
			host = r.RemoteAddr
		}
		if !s.limiter.allow(host) {
			w.Header().Set("Retry-After", strconv.Itoa(int(s.limiter.cfg.Window.Seconds())))
			s.error(w, r, http.StatusTooManyRequests, "rate_limited", "Too many requests.", map[string]any{"retryAfterSeconds": int(s.limiter.cfg.Window.Seconds())})
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
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
