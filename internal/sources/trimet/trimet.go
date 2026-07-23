// Package trimet provides a feature-gated boundary for TriMet Web Services.
// It intentionally has no persistence or HTTP-handler dependencies.
package trimet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const SourceID = "trimet-web-services"

var (
	ErrDisabled       = errors.New("TriMet Web Services adapter is disabled")
	ErrInvalidConfig  = errors.New("invalid TriMet Web Services configuration")
	ErrInvalidRequest = errors.New("invalid TriMet Web Services request")
)

// ErrorKind is safe to expose to application error mapping. It never contains
// a provider body, URL query, or AppID.
type ErrorKind string

const (
	ErrorDisabled    ErrorKind = "disabled"
	ErrorUnavailable ErrorKind = "source_unavailable"
	ErrorTimeout     ErrorKind = "timeout"
	ErrorMalformed   ErrorKind = "malformed_response"
	ErrorInvalid     ErrorKind = "invalid_request"
)

type Error struct {
	Kind       ErrorKind
	StatusCode int
	Err        error
}

func (e *Error) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("TriMet Web Services %s (status %d)", e.Kind, e.StatusCode)
	}
	return fmt.Sprintf("TriMet Web Services %s", e.Kind)
}
func (e *Error) Unwrap() error { return e.Err }

// IsSourceUnavailable tells callers whether to map this to the API's safe
// source_unavailable response. Disabled is deliberately separate: it is a
// feature/configuration state, not a transient provider outage.
func IsSourceUnavailable(err error) bool {
	var sourceErr *Error
	return errors.As(err, &sourceErr) && (sourceErr.Kind == ErrorUnavailable || sourceErr.Kind == ErrorTimeout)
}

type Config struct {
	Enabled        bool
	AppID          string
	BaseURL        string
	AllowedHosts   []string
	Timeout        time.Duration
	PlannerEnabled bool
}

// LoadConfig reads only adapter settings. TRIMET_APP_ID_FILE takes precedence
// over TRIMET_APP_ID, so Compose/Docker secrets can be mounted without placing
// the credential in the environment. Enablement still requires an AppID.
func LoadConfig(getenv func(string) string, readFile func(string) ([]byte, error)) (Config, error) {
	if getenv == nil || readFile == nil {
		return Config{}, fmt.Errorf("%w: environment and file readers are required", ErrInvalidConfig)
	}
	config := Config{
		Enabled:        parseBool(getenv("TRIMET_ENABLED")),
		PlannerEnabled: parseBool(getenv("TRIMET_PLANNER_ENABLED")),
		BaseURL:        strings.TrimSpace(getenv("TRIMET_BASE_URL")),
		AllowedHosts:   splitCSV(getenv("TRIMET_ALLOWED_HOSTS")),
		Timeout:        10 * time.Second,
	}
	if raw := strings.TrimSpace(getenv("TRIMET_TIMEOUT")); raw != "" {
		value, err := time.ParseDuration(raw)
		if err != nil || value <= 0 || value > 60*time.Second {
			return Config{}, fmt.Errorf("%w: TRIMET_TIMEOUT must be between 0 and 60s", ErrInvalidConfig)
		}
		config.Timeout = value
	}
	if path := strings.TrimSpace(getenv("TRIMET_APP_ID_FILE")); path != "" {
		contents, err := readFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("%w: cannot read TRIMET_APP_ID_FILE", ErrInvalidConfig)
		}
		config.AppID = strings.TrimSpace(string(contents))
	} else {
		config.AppID = strings.TrimSpace(getenv("TRIMET_APP_ID"))
	}
	if len(config.AllowedHosts) == 0 {
		config.AllowedHosts = []string{"ws.trimet.org"}
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) Validate() error {
	if c.Timeout <= 0 || c.Timeout > 60*time.Second {
		return fmt.Errorf("%w: timeout must be between 0 and 60s", ErrInvalidConfig)
	}
	if c.PlannerEnabled && !c.Enabled {
		return fmt.Errorf("%w: planner cannot be enabled while source is disabled", ErrInvalidConfig)
	}
	if !c.Enabled {
		return nil
	}
	if c.AppID == "" || strings.ContainsAny(c.AppID, "\r\n") {
		return fmt.Errorf("%w: enabled source requires a non-empty AppID", ErrInvalidConfig)
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || !hostAllowed(u.Hostname(), c.AllowedHosts) {
		return fmt.Errorf("%w: base URL must be HTTPS and match the configured allowlist", ErrInvalidConfig)
	}
	return nil
}

type Clock interface{ Now() time.Time }
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

type Client struct {
	config Config
	http   *http.Client
	clock  Clock
}

func NewClient(config Config, httpClient *http.Client, clock Clock) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: config.Timeout}
	}
	if clock == nil {
		clock = systemClock{}
	}
	return &Client{config: config, http: httpClient, clock: clock}, nil
}

// Freshness preserves provider and fetch timestamps at the source boundary.
type Freshness struct {
	Source          string
	SourceUpdatedAt *time.Time
	EntityUpdatedAt *time.Time
	FetchedAt       time.Time
	ProcessedAt     time.Time
	IsRealtime      bool
}

type Arrivals interface {
	Arrivals(context.Context, ArrivalsRequest) ([]Arrival, Freshness, error)
}
type Network interface {
	Route(context.Context, string) (Route, Freshness, error)
	Stop(context.Context, string) (Stop, Freshness, error)
}
type Vehicles interface {
	Vehicle(context.Context, string) (Vehicle, Freshness, error)
	Trip(context.Context, string) (Trip, Freshness, error)
	Block(context.Context, string) (Block, Freshness, error)
}
type Planner interface {
	Plan(context.Context, PlanRequest) (Plan, Freshness, error)
}

type ArrivalsRequest struct {
	StopID  string
	Minutes int
}
type Arrival struct {
	StopID, RouteID, TripID, VehicleID, Headsign string
	ScheduledAt, EstimatedAt                     *time.Time
	Status                                       string
}
type Route struct{ ID, ShortName, LongName string }
type Stop struct {
	ID, Name            string
	Longitude, Latitude float64
}
type Vehicle struct {
	ID, RouteID, TripID, BlockID string
	Longitude, Latitude          float64
	UpdatedAt                    *time.Time
}
type Trip struct{ ID, RouteID, BlockID string }
type Block struct {
	ID      string
	TripIDs []string
}
type PlanRequest struct {
	Origin, Destination string
	DepartAt            *time.Time
}
type Plan struct {
	ID          string
	Itineraries []Itinerary
}
type Itinerary struct {
	DurationSeconds int
	Transfers       int
}

func (c *Client) Arrivals(ctx context.Context, request ArrivalsRequest) ([]Arrival, Freshness, error) {
	if err := validateID(request.StopID); err != nil {
		return nil, Freshness{}, err
	}
	if request.Minutes < 0 || request.Minutes > 180 {
		return nil, Freshness{}, invalidRequest("minutes must be between 0 and 180")
	}
	var response arrivalsResponse
	freshness, err := c.get(ctx, "/ws/v2/arrivals", url.Values{"locIDs": {request.StopID}, "minutes": {strconv.Itoa(request.Minutes)}}, &response)
	if err != nil {
		return nil, Freshness{}, err
	}
	return mapArrivals(response), freshness, nil
}
func (c *Client) Route(ctx context.Context, id string) (Route, Freshness, error) {
	var v routeResponse
	f, e := c.getID(ctx, "/ws/v2/routeConfig", id, &v)
	return mapRoute(v), f, e
}
func (c *Client) Stop(ctx context.Context, id string) (Stop, Freshness, error) {
	var v stopResponse
	f, e := c.getID(ctx, "/ws/v2/stop", id, &v)
	return mapStop(v), f, e
}
func (c *Client) Vehicle(ctx context.Context, id string) (Vehicle, Freshness, error) {
	var v vehicleResponse
	f, e := c.getID(ctx, "/ws/v2/vehicle", id, &v)
	return mapVehicle(v), f, e
}
func (c *Client) Trip(ctx context.Context, id string) (Trip, Freshness, error) {
	var v tripResponse
	f, e := c.getID(ctx, "/ws/v2/trip", id, &v)
	return mapTrip(v), f, e
}
func (c *Client) Block(ctx context.Context, id string) (Block, Freshness, error) {
	var v blockResponse
	f, e := c.getID(ctx, "/ws/v2/block", id, &v)
	return mapBlock(v), f, e
}
func (c *Client) Plan(ctx context.Context, request PlanRequest) (Plan, Freshness, error) {
	if !c.config.PlannerEnabled {
		return Plan{}, Freshness{}, &Error{Kind: ErrorDisabled, Err: ErrDisabled}
	}
	if err := validateID(request.Origin); err != nil {
		return Plan{}, Freshness{}, err
	}
	if err := validateID(request.Destination); err != nil {
		return Plan{}, Freshness{}, err
	}
	values := url.Values{"fromPlace": {request.Origin}, "toPlace": {request.Destination}}
	if request.DepartAt != nil {
		values.Set("date", request.DepartAt.UTC().Format(time.RFC3339))
	}
	var response planResponse
	f, e := c.get(ctx, "/ws/v2/tripplanner", values, &response)
	return mapPlan(response), f, e
}

func (c *Client) getID(ctx context.Context, path, id string, target any) (Freshness, error) {
	if err := validateID(id); err != nil {
		return Freshness{}, err
	}
	return c.get(ctx, path, url.Values{"id": {id}}, target)
}
func (c *Client) get(ctx context.Context, path string, query url.Values, target any) (Freshness, error) {
	if !c.config.Enabled {
		return Freshness{}, &Error{Kind: ErrorDisabled, Err: ErrDisabled}
	}
	u, _ := url.Parse(c.config.BaseURL)
	u.Path = strings.TrimRight(u.Path, "/") + path
	u.RawQuery = query.Encode()
	u.RawQuery += "&appID=" + url.QueryEscape(c.config.AppID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Freshness{}, &Error{Kind: ErrorInvalid, Err: errors.New("request construction failed")}
	}
	response, err := c.http.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return Freshness{}, &Error{Kind: ErrorTimeout, Err: errors.New("request deadline exceeded")}
		}
		return Freshness{}, &Error{Kind: ErrorUnavailable, Err: errors.New("request failed")}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		return Freshness{}, &Error{Kind: ErrorUnavailable, StatusCode: response.StatusCode}
	}
	if err := decodeResponse(response.Body, target); err != nil {
		return Freshness{}, &Error{Kind: ErrorMalformed, Err: err}
	}
	now := c.clock.Now().UTC()
	return Freshness{Source: SourceID, FetchedAt: now, ProcessedAt: now, IsRealtime: true}, nil
}

func invalidRequest(message string) error {
	return &Error{Kind: ErrorInvalid, Err: fmt.Errorf("%w: %s", ErrInvalidRequest, message)}
}
func validateID(id string) error {
	if id == "" || len(id) > 512 || strings.ContainsAny(id, "\r\n\t") {
		return invalidRequest("identifier is invalid")
	}
	return nil
}
func parseBool(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true") || strings.TrimSpace(value) == "1"
}
func splitCSV(value string) []string {
	var result []string
	for _, v := range strings.Split(value, ",") {
		if v = strings.TrimSpace(strings.ToLower(v)); v != "" {
			result = append(result, v)
		}
	}
	return result
}
func hostAllowed(host string, allowed []string) bool {
	host = strings.ToLower(host)
	for _, candidate := range allowed {
		if host == strings.ToLower(strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}
