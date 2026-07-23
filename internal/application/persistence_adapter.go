package application

import (
	"context"
	"github.com/toozej/tabi-transit/internal/persistence"
)

// PersistenceVehicleStore is the narrow production adapter. Static catalog
// queries are deliberately not guessed until sqlc exposes their final queries.
type PersistenceVehicleStore struct{ Reader persistence.Reader }

func (s PersistenceVehicleStore) ListCurrentVehicles(ctx context.Context, filter persistence.VehicleFilter) ([]persistence.Vehicle, error) {
	return s.Reader.ListCurrentVehicles(ctx, filter)
}
