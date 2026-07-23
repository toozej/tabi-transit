package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/toozej/tabi-transit/internal/sources/gtfs"
	"time"
)

// PostgresStore uses one transaction for version creation, static rows, and
// activation. Any load error rolls back and leaves the previously active feed
// untouched. Public identifiers are the mapping contract consumed by WP-05:
// <source-id>:route:<route_id>, <source-id>:trip:<trip_id>, and
// <source-id>:stop:<stop_id> (shapes use :shape:).
type PostgresStore struct {
	DB interface {
		Begin(context.Context) (pgx.Tx, error)
	}
}

func public(source, kind, id string) string { return source + ":" + kind + ":" + id }
func mode(routeType string) string {
	switch routeType {
	case "0":
		return "light_rail"
	case "2":
		return "commuter_rail"
	case "3":
		return "bus"
	case "4":
		return "ferry"
	case "5":
		return "streetcar"
	}
	return "other"
}
func (p PostgresStore) Import(ctx context.Context, source, label, digest string, fetched time.Time, f gtfs.Feed) (bool, error) {
	tx, e := p.DB.Begin(ctx)
	if e != nil {
		return false, e
	}
	defer tx.Rollback(ctx)
	var id int64
	e = tx.QueryRow(ctx, `SELECT id FROM catalog.feed_versions WHERE source_id=$1 AND archive_sha256=$2`, source, digest).Scan(&id)
	if e == nil {
		var active bool
		e = tx.QueryRow(ctx, `SELECT status='active' FROM catalog.feed_versions WHERE id=$1`, id).Scan(&active)
		if e != nil {
			return false, e
		}
		if active {
			return true, tx.Commit(ctx)
		}
		if _, e = tx.Exec(ctx, `DELETE FROM transit.stop_times WHERE feed_version_id=$1; DELETE FROM transit.trips WHERE feed_version_id=$1; DELETE FROM transit.stops WHERE feed_version_id=$1; DELETE FROM transit.routes WHERE feed_version_id=$1; DELETE FROM transit.services WHERE feed_version_id=$1`, id); e != nil {
			return false, e
		}
	} else if e == pgx.ErrNoRows {
		report, _ := json.Marshal(map[string]int{"stops": len(f.Stops), "routes": len(f.Routes), "trips": len(f.Trips), "stop_times": len(f.StopTimes)})
		e = tx.QueryRow(ctx, `INSERT INTO catalog.feed_versions(source_id,version_label,archive_sha256,fetched_at,import_report,service_timezone) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`, source, label, digest, fetched, report, feedTimezone(f)).Scan(&id)
		if e != nil {
			return false, e
		}
	} else {
		return false, e
	}
	if _, e = tx.Exec(ctx, `UPDATE catalog.feed_versions SET service_timezone=$2 WHERE id=$1`, id, feedTimezone(f)); e != nil {
		return false, e
	}
	for _, s := range f.Stops {
		_, e = tx.Exec(ctx, `INSERT INTO transit.stops(feed_version_id,public_id,source_stop_id,name,parent_public_id,point) VALUES($1,$2,$3,$4,NULLIF($5,''),ST_SetSRID(ST_MakePoint($6,$7),4326)::geography)`, id, public(source, "stop", s.ID), s.ID, s.Name, public(source, "stop", s.ParentID), s.Longitude, s.Latitude)
		if e != nil {
			return false, fmt.Errorf("insert stop: %w", e)
		}
	}
	for _, r := range f.Routes {
		_, e = tx.Exec(ctx, `INSERT INTO transit.routes(feed_version_id,public_id,source_route_id,short_name,long_name,mode) VALUES($1,$2,$3,NULLIF($4,''),NULLIF($5,''),$6::transit.mode)`, id, public(source, "route", r.ID), r.ID, r.ShortName, r.LongName, mode(r.Type))
		if e != nil {
			return false, fmt.Errorf("insert route: %w", e)
		}
	}
	for _, t := range f.Trips {
		_, e = tx.Exec(ctx, `INSERT INTO transit.trips(feed_version_id,public_id,source_trip_id,route_public_id,service_id) VALUES($1,$2,$3,$4,$5)`, id, public(source, "trip", t.ID), t.ID, public(source, "route", t.RouteID), t.ServiceID)
		if e != nil {
			return false, fmt.Errorf("insert trip: %w", e)
		}
	}
	for _, calendar := range f.Calendars {
		_, e = tx.Exec(ctx, `INSERT INTO transit.services(feed_version_id,service_id) VALUES($1,$2)`, id, calendar.ServiceID)
		if e != nil {
			return false, fmt.Errorf("insert service: %w", e)
		}
		_, e = tx.Exec(ctx, `INSERT INTO transit.service_calendars(feed_version_id,service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::date,$11::date)`, id, calendar.ServiceID, calendar.Weekdays[0], calendar.Weekdays[1], calendar.Weekdays[2], calendar.Weekdays[3], calendar.Weekdays[4], calendar.Weekdays[5], calendar.Weekdays[6], calendar.StartDate, calendar.EndDate)
		if e != nil {
			return false, fmt.Errorf("insert service calendar: %w", e)
		}
	}
	for _, exception := range f.Exceptions {
		_, e = tx.Exec(ctx, `INSERT INTO transit.services(feed_version_id,service_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, exception.ServiceID)
		if e != nil {
			return false, fmt.Errorf("insert exception service: %w", e)
		}
		typeValue := 2
		if exception.Added {
			typeValue = 1
		}
		_, e = tx.Exec(ctx, `INSERT INTO transit.service_calendar_dates(feed_version_id,service_id,service_date,exception_type) VALUES($1,$2,$3::date,$4)`, id, exception.ServiceID, exception.Date, typeValue)
		if e != nil {
			return false, fmt.Errorf("insert service calendar date: %w", e)
		}
	}
	for _, st := range f.StopTimes {
		var a, d interface{}
		if st.HasArrival {
			a = st.ArrivalSeconds
		}
		if st.HasDeparture {
			d = st.DepartureSeconds
		}
		_, e = tx.Exec(ctx, `INSERT INTO transit.stop_times(feed_version_id,trip_public_id,stop_public_id,stop_sequence,arrival_seconds,departure_seconds) VALUES($1,$2,$3,$4,$5,$6)`, id, public(source, "trip", st.TripID), public(source, "stop", st.StopID), st.Sequence, a, d)
		if e != nil {
			return false, fmt.Errorf("insert stop_time: %w", e)
		}
	}
	if _, e = tx.Exec(ctx, `UPDATE catalog.feed_versions SET status='superseded' WHERE source_id=$1 AND status='active'; UPDATE catalog.feed_versions SET status='active',activated_at=$2 WHERE id=$3`, source, fetched, id); e != nil {
		return false, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO ops.source_health(source_id,last_attempt_at,last_success_at,entity_count) VALUES($1,$2,$2,$3) ON CONFLICT(source_id) DO UPDATE SET last_attempt_at=$2,last_success_at=$2,consecutive_failures=0,last_error_code=NULL,entity_count=$3,updated_at=now()`, source, fetched, len(f.Stops)); e != nil {
		return false, e
	}
	return false, tx.Commit(ctx)
}

func feedTimezone(f gtfs.Feed) any {
	if len(f.AgencyTimezones) == 0 {
		return nil
	}
	return f.AgencyTimezones[0].Timezone
}
func (p PostgresStore) RecordFailure(ctx context.Context, source, code string, at time.Time) error {
	tx, e := p.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	_, e = tx.Exec(ctx, `INSERT INTO ops.source_health(source_id,last_attempt_at,last_failure_at,consecutive_failures,last_error_code) VALUES($1,$2,$2,1,$3) ON CONFLICT(source_id) DO UPDATE SET last_attempt_at=$2,last_failure_at=$2,consecutive_failures=ops.source_health.consecutive_failures+1,last_error_code=$3,updated_at=now()`, source, at, code)
	if e != nil {
		return e
	}
	return tx.Commit(ctx)
}
