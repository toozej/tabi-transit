package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
	wheelchair := pgtype.Bool{}
	if filter.WheelchairAccessible != nil {
		wheelchair = pgtype.Bool{Bool: *filter.WheelchairAccessible, Valid: true}
	}
	rows, err := r.queries.ListNearbyStopsPerMode(ctx, sqlcgen.ListNearbyStopsPerModeParams{
		LimitPerMode: int64(filter.LimitPerMode), TotalLimit: filter.TotalLimit, Lon: filter.Coordinate.Longitude,
		Lat: filter.Coordinate.Latitude, RadiusMeters: float64(filter.RadiusMeters), Modes: filter.Modes, WheelchairAccessible: wheelchair,
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

func (r *PostgresReader) GetStop(ctx context.Context, id string) (StopDetail, error) {
	row, err := r.queries.GetCatalogStop(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return StopDetail{}, ErrNotFound
	}
	if err != nil {
		return StopDetail{}, err
	}
	routeIDs, err := stringList(row.RoutePublicIds)
	if err != nil {
		return StopDetail{}, err
	}
	stop := catalogStop(row.PublicID, row.Name, string(row.Mode), row.Longitude, row.Latitude, routeIDs, textPtr(row.ParentPublicID), row.WheelchairBoarding)
	return StopDetail{Stop: stop, StaticFeedVersion: row.VersionLabel, Freshness: StaticFreshness{Source: "normalized-static-gtfs", ActivatedAt: timestamp(row.ActivatedAt)}}, nil
}

func (r *PostgresReader) GetRoute(ctx context.Context, id string) (RouteDetail, error) {
	row, err := r.queries.GetCatalogRoute(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return RouteDetail{}, ErrNotFound
	}
	if err != nil {
		return RouteDetail{}, err
	}
	directions, err := r.queries.ListRouteDirections(ctx, id)
	if err != nil {
		return RouteDetail{}, err
	}
	out := make([]RouteDirection, 0, len(directions))
	for _, d := range directions {
		if d.DirectionID.Valid {
			out = append(out, RouteDirection{ID: int(d.DirectionID.Int16), Headsign: d.Headsign})
		}
	}
	return RouteDetail{Route: CatalogRoute{ID: row.PublicID, Mode: string(row.Mode), ShortName: text(row.ShortName), LongName: text(row.LongName), Color: stringPtr(row.Color), TextColor: stringPtr(row.TextColor)}, Directions: out, StaticFeedVersion: row.VersionLabel, Freshness: StaticFreshness{Source: "normalized-static-gtfs", ActivatedAt: timestamp(row.ActivatedAt)}}, nil
}

func (r *PostgresReader) ListRouteStops(ctx context.Context, id string, direction *int) ([]RouteStop, string, error) {
	version, err := r.queries.LatestActiveFeedVersionLabel(ctx)
	if err != nil {
		return nil, "", err
	}
	rows, err := r.queries.ListRouteStops(ctx, sqlcgen.ListRouteStopsParams{RoutePublicID: id, DirectionID: int2(direction)})
	if err != nil {
		return nil, "", err
	}
	out := make([]RouteStop, 0, len(rows))
	for _, row := range rows {
		d := int2Ptr(row.DirectionID)
		out = append(out, RouteStop{Stop: catalogStop(row.PublicID, row.Name, string(row.Mode), row.Longitude, row.Latitude, nil, textPtr(row.ParentPublicID), row.WheelchairBoarding), Sequence: int(row.StopSequence), DirectionID: d})
	}
	return out, version, nil
}

func (r *PostgresReader) ListRouteShapes(ctx context.Context, id string, direction *int) ([]RouteShape, string, error) {
	version, err := r.queries.LatestActiveFeedVersionLabel(ctx)
	if err != nil {
		return nil, "", err
	}
	rows, err := r.queries.ListRouteShapes(ctx, sqlcgen.ListRouteShapesParams{RoutePublicID: id, DirectionID: int2(direction)})
	if err != nil {
		return nil, "", err
	}
	out := make([]RouteShape, 0, len(rows))
	for _, row := range rows {
		var geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		}
		if err := json.Unmarshal([]byte(row.Geometry), &geometry); err != nil {
			return nil, "", fmt.Errorf("route shape %s geometry: %w", row.ShapeID, err)
		}
		out = append(out, RouteShape{ID: row.ShapeID, RouteID: row.RoutePublicID, DirectionID: int2Ptr(row.DirectionID), Coordinates: geometry.Coordinates})
	}
	return out, version, nil
}

func (r *PostgresReader) ListStopSchedule(ctx context.Context, filter ScheduleFilter) ([]ScheduleTime, string, string, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		return nil, "", "", fmt.Errorf("schedule limit out of range")
	}
	date, err := time.Parse("2006-01-02", filter.ServiceDate)
	if err != nil {
		return nil, "", "", err
	}
	rows, err := r.queries.ListStopSchedule(ctx, sqlcgen.ListStopScheduleParams{StopID: filter.StopID, RouteID: filter.RouteID, DirectionID: int2(filter.DirectionID), Cursor: filter.Cursor, RowLimit: int32(filter.Limit + 1), ServiceDate: pgtype.Date{Time: date, Valid: true}})
	if err != nil {
		return nil, "", "", err
	}
	next := ""
	if len(rows) > filter.Limit {
		next = rows[filter.Limit-1].TripID
		rows = rows[:filter.Limit]
	}
	items := make([]ScheduleTime, 0, len(rows))
	version := ""
	for _, row := range rows {
		version = row.VersionLabel
		seconds := int(row.ServiceDaySeconds.Int32)
		// No agency timezone is normalized yet. Keep the authoritative service
		// seconds and omit departureAt rather than label a UTC instant as local.
		items = append(items, ScheduleTime{TripID: row.TripID, RouteID: row.RouteID, StopID: row.StopID, ServiceDate: filter.ServiceDate, DirectionID: int2Ptr(row.DirectionID), Headsign: row.Headsign, ServiceDaySeconds: seconds})
	}
	return items, version, next, nil
}

// ListStopArrivals deliberately remains unavailable until an agency service
// timezone is part of normalized feed metadata. Treating UTC as service time
// would turn valid after-midnight times into false rider-facing estimates.
func (r *PostgresReader) ListStopArrivals(context.Context, ArrivalFilter) ([]Arrival, error) {
	return nil, errors.New("arrival service timezone unavailable")
}

func (r *PostgresReader) ListAlerts(ctx context.Context, filter AlertFilter) ([]Alert, string, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		return nil, "", fmt.Errorf("alert limit out of range")
	}
	updated := pgtype.Timestamptz{}
	if filter.UpdatedSince != nil {
		updated = pgtype.Timestamptz{Time: *filter.UpdatedSince, Valid: true}
	}
	now := filter.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	rows, err := r.queries.ListCurrentAlerts(ctx, sqlcgen.ListCurrentAlertsParams{ActiveOnly: filter.Active, NowAt: pgtype.Timestamptz{Time: now, Valid: true}, Effect: filter.Effect, UpdatedSince: updated, Cursor: filter.Cursor, RowLimit: int32(filter.Limit + 1)})
	if err != nil {
		return nil, "", err
	}
	next := ""
	if len(rows) > filter.Limit {
		last := rows[filter.Limit-1]
		next = last.SourceID + ":alert:" + last.EntityID
		rows = rows[:filter.Limit]
	}
	out := make([]Alert, 0, len(rows))
	for _, row := range rows {
		out = append(out, alertModel(row.SourceID, row.EntityID, textPtr(row.Cause), textPtr(row.Effect), row.HeaderText, row.DescriptionText, row.Url, timestampPtr(row.ActiveFrom), timestampPtr(row.ActiveUntil), timestamp(row.FetchedAt), timestamp(row.ProcessedAt)))
	}
	return out, next, nil
}
func (r *PostgresReader) GetAlert(ctx context.Context, id string) (Alert, error) {
	parts := strings.SplitN(id, ":alert:", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Alert{}, ErrInvalidPublicID
	}
	row, err := r.queries.GetCurrentAlert(ctx, sqlcgen.GetCurrentAlertParams{SourceID: parts[0], EntityID: parts[1]})
	if errors.Is(err, pgx.ErrNoRows) {
		return Alert{}, ErrNotFound
	}
	if err != nil {
		return Alert{}, err
	}
	return alertModel(row.SourceID, row.EntityID, textPtr(row.Cause), textPtr(row.Effect), row.HeaderText, row.DescriptionText, row.Url, timestampPtr(row.ActiveFrom), timestampPtr(row.ActiveUntil), timestamp(row.FetchedAt), timestamp(row.ProcessedAt)), nil
}
func alertModel(source, id string, cause, effect *string, header, description, url string, from, until *time.Time, fetched, processed time.Time) Alert {
	result := Alert{ID: source + ":alert:" + id, Revision: fetched.UTC().Format(time.RFC3339Nano), Header: header, Description: description, SourceURL: url, Source: source, ActiveFrom: from, ActiveUntil: until, FetchedAt: fetched, ProcessedAt: processed}
	if cause != nil {
		result.Cause = *cause
	}
	if effect != nil {
		result.Effect = *effect
	}
	return result
}

func catalogStop(id, name, mode string, longitude, latitude float64, routeIDs []string, parent *string, wheelchair int16) CatalogStop {
	stop := CatalogStop{ID: id, Name: name, Coordinate: Coordinate{Longitude: longitude, Latitude: latitude}, Modes: []string{mode}, RouteIDs: routeIDs, ParentStopID: parent}
	if wheelchair == 1 {
		value := true
		stop.WheelchairAccessible = &value
	} else if wheelchair == 2 {
		value := false
		stop.WheelchairAccessible = &value
	}
	return stop
}
func int2(value *int) pgtype.Int2 {
	if value == nil {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: int16(*value), Valid: true}
}
func int2Ptr(value pgtype.Int2) *int {
	if !value.Valid {
		return nil
	}
	result := int(value.Int16)
	return &result
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
var _ RiderInfoReader = (*PostgresReader)(nil)
