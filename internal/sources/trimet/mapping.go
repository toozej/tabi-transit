package trimet

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The source schemas are deliberately private. These minimal DTOs are backed
// by sanitized fixtures; adding a provider field requires a fixture first.
type arrivalsResponse struct {
	ResultSet struct {
		QueryTime providerTime `json:"queryTime"`
		Arrival   []arrivalDTO `json:"arrival"`
	} `json:"resultSet"`
}
type arrivalDTO struct {
	StopID    providerID   `json:"locid"`
	RouteID   providerID   `json:"route"`
	TripID    string       `json:"tripID"`
	VehicleID string       `json:"vehicleID"`
	Headsign  string       `json:"fullSign"`
	Scheduled providerTime `json:"scheduled"`
	Estimated providerTime `json:"estimated"`
	Status    string       `json:"status"`
	Streetcar bool         `json:"streetCar"`
}

// providerTime accepts the documented Unix epoch milliseconds used by TriMet
// Web Services V2. RFC3339 remains accepted only for sanitized legacy
// fixtures; provider responses are never logged when parsing fails.
type providerTime json.RawMessage

// providerID accepts the string and integer JSON representations emitted by
// TriMet. IDs remain opaque strings at the adapter boundary; fractions,
// booleans, objects, and arrays are rejected as absent rather than coerced.
type providerID json.RawMessage

func (v *providerID) UnmarshalJSON(data []byte) error {
	*v = providerID(append((*v)[:0], data...))
	return nil
}

func (v providerID) String() string {
	raw := strings.TrimSpace(string(v))
	if raw == "" || raw == "null" {
		return ""
	}
	if strings.HasPrefix(raw, `"`) {
		var value string
		if json.Unmarshal(v, &value) == nil {
			return value
		}
		return ""
	}
	if _, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return raw
	}
	return ""
}

func (v *providerTime) UnmarshalJSON(data []byte) error {
	*v = providerTime(append((*v)[:0], data...))
	return nil
}

func (v providerTime) Time() *time.Time {
	raw := strings.Trim(strings.TrimSpace(string(v)), `"`)
	if raw == "" || raw == "null" {
		return nil
	}
	if millis, err := strconv.ParseInt(raw, 10, 64); err == nil {
		parsed := time.UnixMilli(millis).UTC()
		return &parsed
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return &parsed
		}
	}
	return nil
}

type routeResponse struct {
	ResultSet struct {
		Route []routeDTO `json:"route"`
	} `json:"resultSet"`
}
type routeDTO struct {
	ID        string `json:"id"`
	ShortName string `json:"route"`
	LongName  string `json:"desc"`
}
type stopResponse struct {
	ResultSet struct {
		Location []stopDTO `json:"location"`
	} `json:"resultSet"`
}
type stopDTO struct {
	ID        string  `json:"locid"`
	Name      string  `json:"desc"`
	Longitude float64 `json:"lng"`
	Latitude  float64 `json:"lat"`
}
type vehicleResponse struct {
	ResultSet struct {
		Vehicle []vehicleDTO `json:"vehicle"`
	} `json:"resultSet"`
}
type vehicleDTO struct {
	ID        string  `json:"vehicleID"`
	RouteID   string  `json:"route"`
	TripID    string  `json:"tripID"`
	BlockID   string  `json:"blockID"`
	Longitude float64 `json:"lng"`
	Latitude  float64 `json:"lat"`
	Updated   string  `json:"time"`
}
type tripResponse struct {
	ResultSet struct {
		Trip []tripDTO `json:"trip"`
	} `json:"resultSet"`
}
type tripDTO struct {
	ID      string `json:"tripID"`
	RouteID string `json:"route"`
	BlockID string `json:"blockID"`
}
type blockResponse struct {
	ResultSet struct {
		Block []blockDTO `json:"block"`
	} `json:"resultSet"`
}
type blockDTO struct {
	ID      string   `json:"blockID"`
	TripIDs []string `json:"tripIDs"`
}
type planResponse struct {
	ResultSet struct {
		PlanID      string         `json:"planID"`
		Itineraries []itineraryDTO `json:"itineraries"`
	} `json:"resultSet"`
}
type itineraryDTO struct {
	DurationSeconds int      `json:"durationSeconds"`
	Transfers       int      `json:"transfers"`
	Legs            []legDTO `json:"legs"`
}
type legDTO struct {
	Mode           string `json:"mode"`
	RouteID        string `json:"routeID"`
	FromName       string `json:"fromName"`
	ToName         string `json:"toName"`
	Start          string `json:"startTime"`
	End            string `json:"endTime"`
	DistanceMeters *int   `json:"distanceMeters"`
}

func decodeResponse(body io.Reader, target any) error {
	decoder := json.NewDecoder(io.LimitReader(body, 1<<20))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.More() {
		return fmt.Errorf("unexpected trailing JSON value")
	}
	return nil
}

// responseJSONShape returns only JSON object field names and value kinds for
// an opt-in smoke diagnostic. It must never include provider response values.
func responseJSONShape(body []byte) string {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(body, &root); err != nil {
		return "json=invalid"
	}
	parts := []string{"json=object", "root_keys=" + strings.Join(sortedJSONKeys(root), ",")}
	if resultSet, ok := root["resultSet"]; ok {
		var result map[string]json.RawMessage
		if err := json.Unmarshal(resultSet, &result); err != nil {
			parts = append(parts, "resultSet=non_object")
		} else {
			parts = append(parts, "resultSet_keys="+strings.Join(sortedJSONKeys(result), ","))
			if arrivals, ok := result["arrival"]; ok {
				parts = append(parts, firstObjectFieldTypes("arrival", arrivals))
			}
		}
	}
	return strings.Join(parts, " ")
}

func firstObjectFieldTypes(name string, raw json.RawMessage) string {
	var values []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || len(values) == 0 {
		return name + "=not_nonempty_array"
	}
	keys := sortedJSONKeys(values[0])
	if len(keys) > 32 {
		keys = keys[:32]
	}
	types := make([]string, 0, len(keys))
	for _, key := range keys {
		types = append(types, key+":"+jsonValueKind(values[0][key]))
	}
	return name + "_first_types=" + strings.Join(types, ",")
}

func jsonValueKind(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	switch {
	case trimmed == "null":
		return "null"
	case strings.HasPrefix(trimmed, `"`):
		return "string"
	case strings.HasPrefix(trimmed, "{"):
		return "object"
	case strings.HasPrefix(trimmed, "["):
		return "array"
	case trimmed == "true" || trimmed == "false":
		return "boolean"
	default:
		return "number_or_unknown"
	}
}

func sortedJSONKeys(values map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		// Field names are bounded in the diagnostic to avoid copying attacker-
		// controlled response data into logs.
		if len(key) <= 64 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
func mapArrivals(input arrivalsResponse) []Arrival {
	output := make([]Arrival, 0, len(input.ResultSet.Arrival))
	for _, v := range input.ResultSet.Arrival {
		output = append(output, Arrival{StopID: v.StopID.String(), RouteID: v.RouteID.String(), TripID: v.TripID, VehicleID: v.VehicleID, Headsign: v.Headsign, ScheduledAt: v.Scheduled.Time(), EstimatedAt: v.Estimated.Time(), Status: v.Status, Streetcar: v.Streetcar})
	}
	return output
}
func mapRoute(v routeResponse) Route {
	if len(v.ResultSet.Route) == 0 {
		return Route{}
	}
	x := v.ResultSet.Route[0]
	return Route{ID: x.ID, ShortName: x.ShortName, LongName: x.LongName}
}
func mapStop(v stopResponse) Stop {
	if len(v.ResultSet.Location) == 0 {
		return Stop{}
	}
	x := v.ResultSet.Location[0]
	return Stop{ID: x.ID, Name: x.Name, Longitude: x.Longitude, Latitude: x.Latitude}
}
func mapVehicle(v vehicleResponse) Vehicle {
	if len(v.ResultSet.Vehicle) == 0 {
		return Vehicle{}
	}
	x := v.ResultSet.Vehicle[0]
	return Vehicle{ID: x.ID, RouteID: x.RouteID, TripID: x.TripID, BlockID: x.BlockID, Longitude: x.Longitude, Latitude: x.Latitude, UpdatedAt: parseProviderTime(x.Updated)}
}
func mapTrip(v tripResponse) Trip {
	if len(v.ResultSet.Trip) == 0 {
		return Trip{}
	}
	x := v.ResultSet.Trip[0]
	return Trip{ID: x.ID, RouteID: x.RouteID, BlockID: x.BlockID}
}
func mapBlock(v blockResponse) Block {
	if len(v.ResultSet.Block) == 0 {
		return Block{}
	}
	x := v.ResultSet.Block[0]
	return Block{ID: x.ID, TripIDs: x.TripIDs}
}
func mapPlan(v planResponse) Plan {
	output := Plan{ID: v.ResultSet.PlanID, Itineraries: make([]Itinerary, 0, len(v.ResultSet.Itineraries))}
	for _, i := range v.ResultSet.Itineraries {
		itinerary := Itinerary{DurationSeconds: i.DurationSeconds, Transfers: i.Transfers, Legs: make([]ItineraryLeg, 0, len(i.Legs))}
		for _, leg := range i.Legs {
			itinerary.Legs = append(itinerary.Legs, ItineraryLeg{
				Mode: Mode(leg.Mode), RouteID: leg.RouteID, FromName: leg.FromName, ToName: leg.ToName,
				StartAt: parseProviderTime(leg.Start), EndAt: parseProviderTime(leg.End), DistanceMeters: leg.DistanceMeters,
			})
		}
		output.Itineraries = append(output.Itineraries, itinerary)
	}
	return output
}
func parseProviderTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}
	return nil
}
