// Package gtfsrt validates and normalizes GTFS-Realtime vehicle-position feeds.
// Provider protobufs must not cross this boundary.
package gtfsrt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	gtfs "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

var (
	ErrMalformed = errors.New("malformed GTFS-Realtime feed")
	ErrEmpty     = errors.New("unexpectedly empty vehicle snapshot")
)

type Vehicle struct {
	SourceVehicleID string
	RouteID         string
	TripID          string
	Latitude        float64
	Longitude       float64
	EntityUpdatedAt *time.Time
}

type Feed struct {
	SourceUpdatedAt time.Time
	Vehicles        []Vehicle
	DeletedEntities int
	Unmatched       int
	SHA256          string
}
type StopTimeUpdate struct {
	StopSequence          uint32
	StopID                string
	ArrivalDelaySeconds   *int32
	ArrivalTime           *time.Time
	DepartureDelaySeconds *int32
	DepartureTime         *time.Time
	ScheduleRelationship  string
}
type TripUpdate struct {
	EntityID, TripID, RouteID string
	StartDate                 string
	ScheduleRelationship      string
	UpdatedAt                 *time.Time
	StopTimes                 []StopTimeUpdate
}
type TripUpdateFeed struct {
	SourceUpdatedAt time.Time
	Updates         []TripUpdate
	SHA256          string
}
type Alert struct {
	EntityID, Cause, Effect  string
	Header, Description, URL string
	ActiveFrom, ActiveUntil  *time.Time
}

// ParseVehiclePositions accepts a complete protobuf message and rejects feeds
// that cannot safely replace a current vehicle snapshot. Deleted and non-vehicle
// entities are accounted for but are never used as current vehicles.
func ParseVehiclePositions(raw []byte, now time.Time, maxAge, futureSkew time.Duration) (Feed, error) {
	message, updated, err := parseMessage(raw, now, maxAge, futureSkew)
	if err != nil {
		return Feed{}, err
	}
	// A differential feed cannot safely replace a full current snapshot until
	// deletion and merge semantics are implemented end-to-end.
	feed := Feed{SourceUpdatedAt: updated, SHA256: digest(raw)}
	seenVehicles := make(map[string]struct{}, len(message.Entity))
	for _, entity := range message.Entity {
		if entity.GetIsDeleted() {
			feed.DeletedEntities++
			continue
		}
		if entity.Vehicle == nil {
			continue
		}
		position, descriptor := entity.Vehicle.Position, entity.Vehicle.Vehicle
		if descriptor == nil || descriptor.GetId() == "" || position == nil || position.Latitude == nil || position.Longitude == nil {
			return Feed{}, fmt.Errorf("%w: incomplete vehicle entity", ErrMalformed)
		}
		if _, duplicate := seenVehicles[descriptor.GetId()]; duplicate {
			return Feed{}, fmt.Errorf("%w: duplicate vehicle ID", ErrMalformed)
		}
		seenVehicles[descriptor.GetId()] = struct{}{}
		lat, lon := float64(position.GetLatitude()), float64(position.GetLongitude())
		if math.IsNaN(lat) || math.IsNaN(lon) || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
			return Feed{}, fmt.Errorf("%w: impossible vehicle coordinate", ErrMalformed)
		}
		var entityAt *time.Time
		if entity.Vehicle.Timestamp != nil {
			t := time.Unix(int64(entity.Vehicle.GetTimestamp()), 0).UTC()
			if t.After(now.Add(futureSkew)) {
				return Feed{}, fmt.Errorf("%w: future entity timestamp", ErrMalformed)
			}
			entityAt = &t
		}
		vehicle := Vehicle{SourceVehicleID: descriptor.GetId(), Latitude: lat, Longitude: lon, EntityUpdatedAt: entityAt}
		if trip := entity.Vehicle.Trip; trip != nil {
			vehicle.RouteID, vehicle.TripID = trip.GetRouteId(), trip.GetTripId()
		}
		feed.Vehicles = append(feed.Vehicles, vehicle)
	}
	if len(feed.Vehicles) == 0 {
		return Feed{}, ErrEmpty
	}
	return feed, nil
}

// ParseTripUpdates validates a complete, current-state projection. It does not
// apply it to storage; callers must keep the last valid state on any error.
func ParseTripUpdates(raw []byte, now time.Time, maxAge, futureSkew time.Duration) ([]TripUpdate, error) {
	feed, err := ParseTripUpdateFeed(raw, now, maxAge, futureSkew)
	if err != nil {
		return nil, err
	}
	return feed.Updates, nil
}

// ParseTripUpdateFeed retains feed-header freshness separately from per-trip
// timestamps. Both are needed to distinguish a valid estimate from a stale
// current-state projection.
func ParseTripUpdateFeed(raw []byte, now time.Time, maxAge, futureSkew time.Duration) (TripUpdateFeed, error) {
	message, updated, err := parseMessage(raw, now, maxAge, futureSkew)
	if err != nil {
		return TripUpdateFeed{}, err
	}
	seen := map[string]struct{}{}
	updates := []TripUpdate{}
	for _, entity := range message.Entity {
		if entity.GetIsDeleted() || entity.TripUpdate == nil {
			continue
		}
		if entity.GetId() == "" || entity.TripUpdate.Trip == nil || entity.TripUpdate.Trip.GetTripId() == "" {
			return TripUpdateFeed{}, fmt.Errorf("%w: incomplete trip update", ErrMalformed)
		}
		if _, duplicate := seen[entity.GetId()]; duplicate {
			return TripUpdateFeed{}, fmt.Errorf("%w: duplicate trip update", ErrMalformed)
		}
		seen[entity.GetId()] = struct{}{}
		update := TripUpdate{EntityID: entity.GetId(), TripID: entity.TripUpdate.Trip.GetTripId(), RouteID: entity.TripUpdate.Trip.GetRouteId(), StartDate: entity.TripUpdate.Trip.GetStartDate(), ScheduleRelationship: entity.TripUpdate.Trip.GetScheduleRelationship().String()}
		if entity.TripUpdate.Timestamp != nil {
			value := time.Unix(int64(entity.TripUpdate.GetTimestamp()), 0).UTC()
			if value.After(now.Add(futureSkew)) {
				return TripUpdateFeed{}, fmt.Errorf("%w: future trip update", ErrMalformed)
			}
			update.UpdatedAt = &value
		}
		for _, stop := range entity.TripUpdate.StopTimeUpdate {
			if stop.GetStopSequence() == 0 && stop.GetStopId() == "" {
				return TripUpdateFeed{}, fmt.Errorf("%w: trip update stop reference", ErrMalformed)
			}
			entry := StopTimeUpdate{StopSequence: stop.GetStopSequence(), StopID: stop.GetStopId(), ScheduleRelationship: stop.GetScheduleRelationship().String()}
			if stop.Arrival != nil {
				entry.ArrivalDelaySeconds = stop.Arrival.Delay
				entry.ArrivalTime = epochInt64Ptr(stop.Arrival.Time, now, futureSkew)
			}
			if stop.Departure != nil {
				entry.DepartureDelaySeconds = stop.Departure.Delay
				entry.DepartureTime = epochInt64Ptr(stop.Departure.Time, now, futureSkew)
			}
			if (stop.Arrival != nil && stop.Arrival.Time != nil && entry.ArrivalTime == nil) || (stop.Departure != nil && stop.Departure.Time != nil && entry.DepartureTime == nil) {
				return TripUpdateFeed{}, fmt.Errorf("%w: future stop update", ErrMalformed)
			}
			update.StopTimes = append(update.StopTimes, entry)
		}
		updates = append(updates, update)
	}
	if len(updates) == 0 {
		return TripUpdateFeed{}, ErrEmpty
	}
	return TripUpdateFeed{SourceUpdatedAt: updated, Updates: updates, SHA256: digest(raw)}, nil
}

// ParseAlerts preserves only bounded rider-facing fields from current alerts.
func ParseAlerts(raw []byte, now time.Time, maxAge, futureSkew time.Duration) ([]Alert, error) {
	message, _, err := parseMessage(raw, now, maxAge, futureSkew)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	alerts := []Alert{}
	for _, entity := range message.Entity {
		if entity.GetIsDeleted() || entity.Alert == nil {
			continue
		}
		if entity.GetId() == "" {
			return nil, fmt.Errorf("%w: alert ID", ErrMalformed)
		}
		if _, duplicate := seen[entity.GetId()]; duplicate {
			return nil, fmt.Errorf("%w: duplicate alert", ErrMalformed)
		}
		seen[entity.GetId()] = struct{}{}
		alert := Alert{EntityID: entity.GetId(), Cause: entity.Alert.GetCause().String(), Effect: entity.Alert.GetEffect().String(), Header: translated(entity.Alert.HeaderText), Description: translated(entity.Alert.DescriptionText), URL: translated(entity.Alert.Url)}
		if len(alert.Header) > 4096 || len(alert.Description) > 16384 || len(alert.URL) > 2048 {
			return nil, fmt.Errorf("%w: alert text too long", ErrMalformed)
		}
		for _, period := range entity.Alert.ActivePeriod {
			// Alert active windows are allowed to extend into the future; the feed
			// header remains the freshness authority for the current projection.
			from, until := epochAnyPtr(period.Start), epochAnyPtr(period.End)
			if from != nil && until != nil && until.Before(*from) {
				return nil, fmt.Errorf("%w: invalid alert period", ErrMalformed)
			}
			if alert.ActiveFrom == nil || (from != nil && from.Before(*alert.ActiveFrom)) {
				alert.ActiveFrom = from
			}
			if alert.ActiveUntil == nil || (until != nil && until.After(*alert.ActiveUntil)) {
				alert.ActiveUntil = until
			}
		}
		alerts = append(alerts, alert)
	}
	if len(alerts) == 0 {
		return nil, ErrEmpty
	}
	return alerts, nil
}
func parseMessage(raw []byte, now time.Time, maxAge, futureSkew time.Duration) (*gtfs.FeedMessage, time.Time, error) {
	if len(raw) == 0 {
		return nil, time.Time{}, fmt.Errorf("%w: empty payload", ErrMalformed)
	}
	var message gtfs.FeedMessage
	if err := proto.Unmarshal(raw, &message); err != nil {
		return nil, time.Time{}, fmt.Errorf("%w: protobuf decode", ErrMalformed)
	}
	if message.Header == nil || message.Header.GetGtfsRealtimeVersion() == "" || message.Header.Timestamp == nil {
		return nil, time.Time{}, fmt.Errorf("%w: required header missing", ErrMalformed)
	}
	if message.Header.GetIncrementality() == gtfs.FeedHeader_DIFFERENTIAL {
		return nil, time.Time{}, fmt.Errorf("%w: differential feeds are not supported", ErrMalformed)
	}
	updated := time.Unix(int64(message.Header.GetTimestamp()), 0).UTC()
	if updated.After(now.Add(futureSkew)) || (maxAge > 0 && now.Sub(updated) > maxAge) {
		return nil, time.Time{}, fmt.Errorf("%w: feed timestamp outside permitted age", ErrMalformed)
	}
	return &message, updated, nil
}
func epochPtr(value *uint64, now time.Time, futureSkew time.Duration) *time.Time {
	if value == nil {
		return nil
	}
	result := time.Unix(int64(*value), 0).UTC()
	if result.After(now.Add(futureSkew)) {
		return nil
	}
	return &result
}
func epochInt64Ptr(value *int64, now time.Time, futureSkew time.Duration) *time.Time {
	if value == nil {
		return nil
	}
	result := time.Unix(*value, 0).UTC()
	if result.After(now.Add(futureSkew)) {
		return nil
	}
	return &result
}
func epochAnyPtr(value *uint64) *time.Time {
	if value == nil {
		return nil
	}
	result := time.Unix(int64(*value), 0).UTC()
	return &result
}
func translated(text *gtfs.TranslatedString) string {
	if text == nil || len(text.Translation) == 0 {
		return ""
	}
	return text.Translation[0].GetText()
}

func digest(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
