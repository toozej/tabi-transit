package transitdataspike

import (
	"fmt"

	gtfs "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

type RealtimeVehicle struct {
	ID                  string
	Latitude, Longitude float32
	Timestamp           uint64
}

// ParseVehiclePositions unmarshals an official GTFS-Realtime protobuf message
// and applies only basic spike-level structural and coordinate validation.
func ParseVehiclePositions(bytes []byte) ([]RealtimeVehicle, error) {
	var feed gtfs.FeedMessage
	if err := proto.Unmarshal(bytes, &feed); err != nil {
		return nil, fmt.Errorf("decode GTFS-Realtime protobuf: %w", err)
	}
	if feed.Header == nil || feed.Header.GtfsRealtimeVersion == nil || feed.Header.GetGtfsRealtimeVersion() == "" {
		return nil, fmt.Errorf("GTFS-Realtime header version is required")
	}
	if feed.Header.Timestamp == nil {
		return nil, fmt.Errorf("GTFS-Realtime header timestamp is required")
	}
	vehicles := make([]RealtimeVehicle, 0)
	for _, entity := range feed.Entity {
		if entity.GetIsDeleted() || entity.Vehicle == nil {
			continue
		}
		position, vehicle := entity.Vehicle.Position, entity.Vehicle.Vehicle
		if entity.Id == nil || entity.GetId() == "" || position == nil || vehicle == nil || vehicle.Id == nil || vehicle.GetId() == "" {
			return nil, fmt.Errorf("vehicle entity is incomplete")
		}
		if position.Latitude == nil || position.Longitude == nil || position.GetLatitude() < -90 || position.GetLatitude() > 90 || position.GetLongitude() < -180 || position.GetLongitude() > 180 {
			return nil, fmt.Errorf("vehicle %s has impossible coordinate", vehicle.GetId())
		}
		vehicles = append(vehicles, RealtimeVehicle{ID: vehicle.GetId(), Latitude: position.GetLatitude(), Longitude: position.GetLongitude(), Timestamp: entity.Vehicle.GetTimestamp()})
	}
	if len(vehicles) == 0 {
		return nil, fmt.Errorf("unexpectedly empty vehicle snapshot")
	}
	return vehicles, nil
}
