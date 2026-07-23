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

// ParseVehiclePositions accepts a complete protobuf message and rejects feeds
// that cannot safely replace a current vehicle snapshot. Deleted and non-vehicle
// entities are accounted for but are never used as current vehicles.
func ParseVehiclePositions(raw []byte, now time.Time, maxAge, futureSkew time.Duration) (Feed, error) {
	if len(raw) == 0 {
		return Feed{}, fmt.Errorf("%w: empty payload", ErrMalformed)
	}
	var message gtfs.FeedMessage
	if err := proto.Unmarshal(raw, &message); err != nil {
		return Feed{}, fmt.Errorf("%w: protobuf decode", ErrMalformed)
	}
	if message.Header == nil || message.Header.GetGtfsRealtimeVersion() == "" || message.Header.Timestamp == nil {
		return Feed{}, fmt.Errorf("%w: required header missing", ErrMalformed)
	}
	// A differential feed cannot safely replace a full current snapshot until
	// deletion and merge semantics are implemented end-to-end.
	if message.Header.GetIncrementality() == gtfs.FeedHeader_DIFFERENTIAL {
		return Feed{}, fmt.Errorf("%w: differential feeds are not supported", ErrMalformed)
	}
	updated := time.Unix(int64(message.Header.GetTimestamp()), 0).UTC()
	if updated.After(now.Add(futureSkew)) || (maxAge > 0 && now.Sub(updated) > maxAge) {
		return Feed{}, fmt.Errorf("%w: feed timestamp outside permitted age", ErrMalformed)
	}
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

func digest(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }
