package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// PostgresRealtimeWriter replaces bounded current-state projections. Each
// successful feed is inserted and swapped inside one transaction; callers
// never use this type for malformed or unexpectedly empty payloads.
type PostgresRealtimeWriter struct {
	DB interface {
		Begin(context.Context) (pgx.Tx, error)
	}
}

func (w PostgresRealtimeWriter) ReplaceVehicleSnapshot(ctx context.Context, snapshot VehicleSnapshot) error {
	if snapshot.SourceID == "" || len(snapshot.Vehicles) == 0 || len(snapshot.ContentSHA256) != 64 {
		return errors.New("valid non-empty vehicle snapshot is required")
	}
	if snapshot.FetchedAt.IsZero() || snapshot.ProcessedAt.IsZero() {
		return errors.New("vehicle snapshot timestamps are required")
	}
	tx, err := w.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var snapshotID int64
	err = tx.QueryRow(ctx, `INSERT INTO realtime.snapshots(source_id,source_updated_at,fetched_at,processed_at,entity_count,content_sha256,is_valid,validation_report)
VALUES($1,$2,$3,$4,$5,$6,true,jsonb_build_object('projection','vehicles'))
ON CONFLICT(source_id,content_sha256) DO UPDATE SET fetched_at=EXCLUDED.fetched_at,processed_at=EXCLUDED.processed_at,is_valid=true
RETURNING id`, snapshot.SourceID, snapshot.SourceUpdatedAt, snapshot.FetchedAt, snapshot.ProcessedAt, len(snapshot.Vehicles), snapshot.ContentSHA256).Scan(&snapshotID)
	if err != nil {
		return fmt.Errorf("create vehicle snapshot: %w", err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM realtime.vehicle_current WHERE source_id=$1`, snapshot.SourceID); err != nil {
		return fmt.Errorf("clear vehicles: %w", err)
	}
	for _, vehicle := range snapshot.Vehicles {
		if vehicle.SourceID != snapshot.SourceID || ValidatePublicID(vehicle.ID, "vehicle") != nil || vehicle.SourceVehicleID == "" || vehicle.Coordinate.Latitude < -90 || vehicle.Coordinate.Latitude > 90 || vehicle.Coordinate.Longitude < -180 || vehicle.Coordinate.Longitude > 180 {
			return errors.New("invalid vehicle current-state record")
		}
		if _, err = tx.Exec(ctx, `INSERT INTO realtime.vehicle_current(source_id,public_id,source_vehicle_id,snapshot_id,route_public_id,trip_public_id,mode,point,source_updated_at,entity_updated_at,fetched_at,processed_at,freshness_status)
VALUES($1,$2,$3,$4,$5,$6,$7::transit.mode,ST_SetSRID(ST_MakePoint($8,$9),4326)::geography,$10,$11,$12,$13,$14::realtime.freshness_status)`, snapshot.SourceID, vehicle.ID, vehicle.SourceVehicleID, snapshotID, vehicle.RouteID, vehicle.TripID, vehicle.Mode, vehicle.Coordinate.Longitude, vehicle.Coordinate.Latitude, vehicle.SourceUpdatedAt, vehicle.EntityUpdatedAt, vehicle.FetchedAt, vehicle.ProcessedAt, vehicle.Freshness); err != nil {
			return fmt.Errorf("insert vehicle: %w", err)
		}
	}
	if err = w.recordSuccess(ctx, tx, snapshot.SourceID, snapshot.FetchedAt, snapshot.SourceUpdatedAt, len(snapshot.Vehicles)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (w PostgresRealtimeWriter) ReplaceTripUpdateSnapshot(ctx context.Context, snapshot TripUpdateSnapshot) error {
	if snapshot.SourceID == "" || len(snapshot.Updates) == 0 || snapshot.ContentSHA256 == "" {
		return errors.New("valid non-empty trip update snapshot is required")
	}
	if snapshot.FetchedAt.IsZero() || snapshot.ProcessedAt.IsZero() {
		return errors.New("trip update snapshot timestamps are required")
	}
	tx, err := w.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var snapshotID int64
	err = tx.QueryRow(ctx, `INSERT INTO realtime.snapshots(source_id,source_updated_at,fetched_at,processed_at,entity_count,content_sha256,is_valid,validation_report)
VALUES($1,$2,$3,$4,$5,$6,true,jsonb_build_object('projection','trip_updates'))
ON CONFLICT(source_id,content_sha256) DO UPDATE SET fetched_at=EXCLUDED.fetched_at,processed_at=EXCLUDED.processed_at,is_valid=true
RETURNING id`, snapshot.SourceID, snapshot.SourceUpdatedAt, snapshot.FetchedAt, snapshot.ProcessedAt, len(snapshot.Updates), snapshot.ContentSHA256).Scan(&snapshotID)
	if err != nil {
		return fmt.Errorf("create trip update snapshot: %w", err)
	}
	// Delete only after the replacement snapshot row has been accepted. Any
	// later failure rolls this transaction back, preserving the prior current
	// projection exactly.
	if _, err = tx.Exec(ctx, `DELETE FROM realtime.trip_updates_current WHERE source_id=$1`, snapshot.SourceID); err != nil {
		return fmt.Errorf("clear trip updates: %w", err)
	}
	for _, update := range snapshot.Updates {
		if update.EntityID == "" || update.TripID == "" {
			return errors.New("trip update entity and trip IDs are required")
		}
		var startDate any
		if update.StartDate != "" {
			parsed, parseErr := time.Parse("20060102", update.StartDate)
			if parseErr != nil {
				return fmt.Errorf("trip update start date: %w", parseErr)
			}
			startDate = parsed
		}
		if _, err = tx.Exec(ctx, `INSERT INTO realtime.trip_updates_current(source_id,entity_id,snapshot_id,trip_public_id,route_public_id,start_date,schedule_relationship,source_updated_at,fetched_at,processed_at)
VALUES($1,$2,$3,$4,NULLIF($5,''),$6,NULLIF($7,''),$8,$9,$10)`, snapshot.SourceID, update.EntityID, snapshotID, update.TripID, update.RouteID, startDate, update.ScheduleRelationship, update.UpdatedAt, snapshot.FetchedAt, snapshot.ProcessedAt); err != nil {
			return fmt.Errorf("insert trip update: %w", err)
		}
		for _, stop := range update.StopTimes {
			if stop.StopSequence < 1 && stop.StopID == "" {
				return errors.New("trip update stop reference is required")
			}
			if _, err = tx.Exec(ctx, `INSERT INTO realtime.trip_update_stop_times_current(source_id,entity_id,stop_sequence,stop_public_id,arrival_delay_seconds,arrival_time,departure_delay_seconds,departure_time,schedule_relationship)
VALUES($1,$2,$3,NULLIF($4,''),$5,$6,$7,$8,NULLIF($9,''))`, snapshot.SourceID, update.EntityID, stop.StopSequence, stop.StopID, stop.ArrivalDelaySeconds, stop.ArrivalTime, stop.DepartureDelaySeconds, stop.DepartureTime, stop.ScheduleRelationship); err != nil {
				return fmt.Errorf("insert trip update stop time: %w", err)
			}
		}
	}
	if err = w.recordSuccess(ctx, tx, snapshot.SourceID, snapshot.FetchedAt, snapshot.SourceUpdatedAt, len(snapshot.Updates)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (w PostgresRealtimeWriter) RecordSourceFailure(ctx context.Context, sourceID, safeCode string, attemptedAt time.Time) error {
	if sourceID == "" || safeCode == "" || len(safeCode) > 128 {
		return errors.New("invalid safe source failure")
	}
	tx, err := w.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `INSERT INTO ops.source_health(source_id,last_attempt_at,last_failure_at,consecutive_failures,last_error_code)
VALUES($1,$2,$2,1,$3)
ON CONFLICT(source_id) DO UPDATE SET last_attempt_at=EXCLUDED.last_attempt_at,last_failure_at=EXCLUDED.last_failure_at,consecutive_failures=ops.source_health.consecutive_failures+1,last_error_code=EXCLUDED.last_error_code,updated_at=now()`, sourceID, attemptedAt, safeCode); err != nil {
		return fmt.Errorf("record source failure: %w", err)
	}
	return tx.Commit(ctx)
}

func (w PostgresRealtimeWriter) recordSuccess(ctx context.Context, tx pgx.Tx, sourceID string, fetchedAt time.Time, sourceUpdatedAt *time.Time, entityCount int) error {
	if _, err := tx.Exec(ctx, `INSERT INTO ops.source_health(source_id,last_attempt_at,last_success_at,last_source_updated_at,last_valid_snapshot_at,consecutive_failures,last_error_code,entity_count)
VALUES($1,$2,$2,$3,$2,0,NULL,$4)
ON CONFLICT(source_id) DO UPDATE SET last_attempt_at=EXCLUDED.last_attempt_at,last_success_at=EXCLUDED.last_success_at,last_source_updated_at=EXCLUDED.last_source_updated_at,last_valid_snapshot_at=EXCLUDED.last_valid_snapshot_at,consecutive_failures=0,last_error_code=NULL,entity_count=EXCLUDED.entity_count,updated_at=now()`, sourceID, fetchedAt, sourceUpdatedAt, entityCount); err != nil {
		return fmt.Errorf("update source health: %w", err)
	}
	return nil
}
