package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/toozej/tabi-transit/internal/persistence/sqlcgen"
)

// PostgresReader is the read-only pgx/sqlc implementation used by HTTP
// composition. It contains no transaction ownership and only returns Tabi's
// normalized persistence models.
type PostgresReader struct{ queries sqlcgen.Querier }

func NewPostgresReader(queries sqlcgen.Querier) *PostgresReader {
	return &PostgresReader{queries: queries}
}

func (r *PostgresReader) ActiveFeedVersion(ctx context.Context, sourceID string) (FeedVersion, error) {
	row, err := r.queries.GetActiveFeedVersion(ctx, sourceID)
	if err != nil {
		return FeedVersion{}, err
	}
	return FeedVersion{ID: row.ID, SourceID: row.SourceID, VersionLabel: row.VersionLabel, ArchiveSHA256: row.ArchiveSha256, ActivatedAt: timestamp(row.ActivatedAt)}, nil
}

func (r *PostgresReader) SourceHealth(ctx context.Context, sourceID string) (SourceHealth, error) {
	row, err := r.queries.GetSourceHealth(ctx, sourceID)
	if err != nil {
		return SourceHealth{}, err
	}
	return SourceHealth{
		SourceID: row.SourceID, LastAttemptAt: timestampPtr(row.LastAttemptAt), LastSuccessAt: timestampPtr(row.LastSuccessAt),
		LastValidSnapshotAt: timestampPtr(row.LastValidSnapshotAt), ConsecutiveFailures: row.ConsecutiveFailures,
		LastErrorCode: textPtr(row.LastErrorCode), EntityCount: int32Ptr(row.EntityCount),
	}, nil
}

func (r *PostgresReader) ListCurrentVehicles(ctx context.Context, filter VehicleFilter) ([]Vehicle, error) {
	rows, err := r.queries.ListCurrentVehicles(ctx, filter.SourceIDs)
	if err != nil {
		return nil, err
	}
	vehicles := make([]Vehicle, 0, len(rows))
	for _, row := range rows {
		vehicles = append(vehicles, Vehicle{
			ID: row.PublicID, SourceID: row.SourceID, SourceVehicleID: row.SourceVehicleID,
			RouteID: textPtr(row.RoutePublicID), TripID: textPtr(row.TripPublicID), Mode: string(row.Mode),
			Coordinate:      Coordinate{Longitude: row.Longitude, Latitude: row.Latitude},
			SourceUpdatedAt: timestampPtr(row.SourceUpdatedAt), EntityUpdatedAt: timestampPtr(row.EntityUpdatedAt),
			FetchedAt: timestamp(row.FetchedAt), ProcessedAt: timestamp(row.ProcessedAt),
			Freshness: FreshnessStatus(row.FreshnessStatus), SnapshotID: row.SnapshotID,
		})
	}
	return vehicles, nil
}

func (r *PostgresReader) ListNearbyStops(ctx context.Context, filter NearbyStopsFilter) ([]NearbyStop, error) {
	rows, err := r.queries.ListNearbyStopsPerMode(ctx, sqlcgen.ListNearbyStopsPerModeParams{
		LimitPerMode: int64(filter.LimitPerMode), TotalLimit: filter.TotalLimit, Lon: filter.Coordinate.Longitude,
		Lat: filter.Coordinate.Latitude, FeedVersionID: filter.FeedVersionID, RadiusMeters: float64(filter.RadiusMeters),
	})
	if err != nil {
		return nil, err
	}
	stops := make([]NearbyStop, 0, len(rows))
	for _, row := range rows {
		lon, err := floatValue(row.Longitude)
		if err != nil {
			return nil, fmt.Errorf("nearby stop %s longitude: %w", row.PublicID, err)
		}
		lat, err := floatValue(row.Latitude)
		if err != nil {
			return nil, fmt.Errorf("nearby stop %s latitude: %w", row.PublicID, err)
		}
		stops = append(stops, NearbyStop{ID: row.PublicID, Name: row.Name, Mode: string(row.Mode), Coordinate: Coordinate{Longitude: lon, Latitude: lat}, DistanceMeters: row.DistanceMeters})
	}
	return stops, nil
}

func (r *PostgresReader) ListCatalogRoutes(ctx context.Context, filter CatalogFilter) (CatalogPage[CatalogRoute], error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		return CatalogPage[CatalogRoute]{}, fmt.Errorf("catalog route limit out of range")
	}
	version, err := r.queries.LatestActiveFeedVersionLabel(ctx)
	if err != nil {
		return CatalogPage[CatalogRoute]{}, err
	}
	rows, err := r.queries.ListRoutes(ctx, sqlcgen.ListRoutesParams{Modes: filter.Modes, Query: filter.Query, Cursor: filter.Cursor, RowLimit: int32(filter.Limit + 1)})
	if err != nil {
		return CatalogPage[CatalogRoute]{}, err
	}
	page := CatalogPage[CatalogRoute]{StaticFeedVersion: version}
	if len(rows) > filter.Limit {
		page.NextCursor = rows[filter.Limit-1].PublicID
		rows = rows[:filter.Limit]
	}
	page.Items = make([]CatalogRoute, 0, len(rows))
	for _, row := range rows {
		page.Items = append(page.Items, CatalogRoute{ID: row.PublicID, Mode: string(row.Mode), ShortName: text(row.ShortName), LongName: text(row.LongName), Color: stringPtr(row.Color), TextColor: stringPtr(row.TextColor)})
	}
	return page, nil
}

func (r *PostgresReader) ListCatalogStops(ctx context.Context, filter CatalogFilter) (CatalogPage[CatalogStop], error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		return CatalogPage[CatalogStop]{}, fmt.Errorf("catalog stop limit out of range")
	}
	version, err := r.queries.LatestActiveFeedVersionLabel(ctx)
	if err != nil {
		return CatalogPage[CatalogStop]{}, err
	}
	rows, err := r.queries.ListStops(ctx, sqlcgen.ListStopsParams{Modes: filter.Modes, Query: filter.Query, Cursor: filter.Cursor, RowLimit: int32(filter.Limit + 1)})
	if err != nil {
		return CatalogPage[CatalogStop]{}, err
	}
	page := CatalogPage[CatalogStop]{StaticFeedVersion: version}
	if len(rows) > filter.Limit {
		page.NextCursor = rows[filter.Limit-1].PublicID
		rows = rows[:filter.Limit]
	}
	page.Items = make([]CatalogStop, 0, len(rows))
	for _, row := range rows {
		routeIDs, err := stringList(row.RoutePublicIds)
		if err != nil {
			return CatalogPage[CatalogStop]{}, fmt.Errorf("stop %s route IDs: %w", row.PublicID, err)
		}
		stop := CatalogStop{ID: row.PublicID, Name: row.Name, Coordinate: Coordinate{Longitude: row.Longitude, Latitude: row.Latitude}, Modes: []string{string(row.Mode)}, RouteIDs: routeIDs, ParentStopID: textPtr(row.ParentPublicID)}
		switch row.WheelchairBoarding {
		case 1:
			value := true
			stop.WheelchairAccessible = &value
		case 2:
			value := false
			stop.WheelchairAccessible = &value
		}
		page.Items = append(page.Items, stop)
	}
	return page, nil
}

// Ready reports whether the API has current, valid vehicle data, not merely a
// reachable database. Callers compose connection health separately.
func (r *PostgresReader) Ready(ctx context.Context) error {
	ready, err := r.queries.HasReadyVehicleData(ctx)
	if err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("no valid vehicle snapshot")
	}
	return nil
}

func timestamp(value pgtype.Timestamptz) time.Time { return value.Time }
func timestampPtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}
func text(value pgtype.Text) string { return value.String }
func textPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}
func int32Ptr(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	result := value.Int32
	return &result
}
func stringPtr(value any) *string {
	text, err := valueString(value)
	if err != nil || text == "" {
		return nil
	}
	return &text
}
func stringList(value any) ([]string, error) {
	raw, err := valueString(value)
	if err != nil {
		return nil, err
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}
func valueString(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	default:
		return "", fmt.Errorf("unexpected database string type %T", value)
	}
}
func floatValue(value any) (float64, error) {
	switch typed := value.(type) {
	case float64:
		return typed, nil
	case float32:
		return float64(typed), nil
	case string:
		return strconv.ParseFloat(typed, 64)
	case []byte:
		return strconv.ParseFloat(string(typed), 64)
	default:
		return 0, fmt.Errorf("unexpected database numeric type %T", value)
	}
}

var _ Reader = (*PostgresReader)(nil)
var _ CatalogReader = (*PostgresReader)(nil)
