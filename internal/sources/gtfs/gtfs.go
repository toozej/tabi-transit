// Package gtfs parses and validates static GTFS archives without making any
// network request.  Provider access is deliberately a boundary owned by the
// importer, so parsing remains deterministic and fixture-testable.
package gtfs

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrArchiveTooLarge = errors.New("gtfs archive exceeds configured limit")
	ErrUnsafeArchive   = errors.New("unsafe GTFS archive")
	ErrInvalidFeed     = errors.New("invalid GTFS feed")
)

type ArchivePolicy struct {
	MaxBytes            int64
	MaxFiles            int
	MaxExpandedBytes    int64
	MaxCompressionRatio float64
}

func DefaultArchivePolicy() ArchivePolicy {
	return ArchivePolicy{MaxBytes: 100 << 20, MaxFiles: 100, MaxExpandedBytes: 500 << 20, MaxCompressionRatio: 100}
}

type Stop struct {
	ID, Name, ParentID  string
	Longitude, Latitude float64
}
type Route struct{ ID, Type, ShortName, LongName string }
type Trip struct{ ID, RouteID, ServiceID, ShapeID string }
type StopTime struct {
	TripID, StopID                             string
	Sequence, ArrivalSeconds, DepartureSeconds int
	HasArrival, HasDeparture                   bool
}
type Calendar struct {
	ServiceID          string
	StartDate, EndDate string
	Weekdays           [7]bool // Monday through Sunday, as defined by GTFS.
}
type CalendarException struct {
	ServiceID string
	Date      string
	Added     bool // true is exception_type=1; false is exception_type=2.
}
type Feed struct {
	Stops      []Stop
	Routes     []Route
	Trips      []Trip
	StopTimes  []StopTime
	Calendars  []Calendar
	Exceptions []CalendarException
	SHA256     string
}

// Read validates archive topology before decoding required CSV files.
func Read(r io.Reader, policy ArchivePolicy) (Feed, error) {
	if policy.MaxBytes <= 0 || policy.MaxFiles <= 0 || policy.MaxExpandedBytes <= 0 || policy.MaxCompressionRatio <= 0 {
		return Feed{}, fmt.Errorf("%w: invalid archive policy", ErrUnsafeArchive)
	}
	b, err := io.ReadAll(io.LimitReader(r, policy.MaxBytes+1))
	if err != nil {
		return Feed{}, err
	}
	if int64(len(b)) > policy.MaxBytes {
		return Feed{}, ErrArchiveTooLarge
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return Feed{}, fmt.Errorf("%w: %v", ErrInvalidFeed, err)
	}
	if len(zr.File) > policy.MaxFiles {
		return Feed{}, fmt.Errorf("%w: too many files", ErrUnsafeArchive)
	}
	files := map[string][]map[string]string{}
	var expanded uint64
	for _, f := range zr.File {
		name := path.Clean(f.Name)
		if name == "." || strings.HasPrefix(name, "../") || path.IsAbs(f.Name) || strings.Contains(f.Name, "\\") {
			return Feed{}, fmt.Errorf("%w: path", ErrUnsafeArchive)
		}
		if _, ok := files[name]; ok {
			return Feed{}, fmt.Errorf("%w: duplicate %s", ErrUnsafeArchive, name)
		}
		if f.UncompressedSize64 > uint64(policy.MaxExpandedBytes) || f.UncompressedSize64 > uint64(float64(maxInt64(f.CompressedSize64, 1))*policy.MaxCompressionRatio) {
			return Feed{}, fmt.Errorf("%w: expansion", ErrUnsafeArchive)
		}
		expanded += f.UncompressedSize64
		if expanded > uint64(policy.MaxExpandedBytes) {
			return Feed{}, fmt.Errorf("%w: expanded bytes", ErrUnsafeArchive)
		}
		if !strings.HasSuffix(name, ".txt") {
			continue
		}
		rows, err := csvRows(f)
		if err != nil {
			return Feed{}, err
		}
		files[name] = rows
	}
	for _, required := range []string{"stops.txt", "routes.txt", "trips.txt", "stop_times.txt"} {
		if _, ok := files[required]; !ok {
			return Feed{}, fmt.Errorf("%w: missing %s", ErrInvalidFeed, required)
		}
	}
	feed, err := build(files)
	if err != nil {
		return Feed{}, err
	}
	sum := sha256.Sum256(b)
	feed.SHA256 = hex.EncodeToString(sum[:])
	return feed, nil
}
func maxInt64(v uint64, min uint64) uint64 {
	if v < min {
		return min
	}
	return v
}
func csvRows(f *zip.File) ([]map[string]string, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	r := csv.NewReader(io.LimitReader(rc, int64(f.UncompressedSize64)+1))
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("%w: %s header", ErrInvalidFeed, f.Name)
	}
	out := []map[string]string{}
	for {
		rec, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, fmt.Errorf("%w: %s CSV", ErrInvalidFeed, f.Name)
		}
		if len(rec) != len(header) {
			return nil, fmt.Errorf("%w: %s row width", ErrInvalidFeed, f.Name)
		}
		row := map[string]string{}
		for i, k := range header {
			row[strings.TrimSpace(k)] = strings.TrimSpace(rec[i])
		}
		out = append(out, row)
	}
	return out, nil
}
func need(row map[string]string, key, table string) (string, error) {
	v := row[key]
	if v == "" {
		return "", fmt.Errorf("%w: %s.%s", ErrInvalidFeed, table, key)
	}
	return v, nil
}
func build(t map[string][]map[string]string) (Feed, error) {
	f := Feed{}
	stops := map[string]bool{}
	routes := map[string]bool{}
	trips := map[string]bool{}
	services := map[string]bool{}
	for _, r := range t["stops.txt"] {
		id, e := need(r, "stop_id", "stops")
		if e != nil {
			return f, e
		}
		if stops[id] {
			return f, fmt.Errorf("%w: duplicate stop", ErrInvalidFeed)
		}
		lon, e := number(r, "stop_lon", "stops")
		if e != nil {
			return f, e
		}
		lat, e := number(r, "stop_lat", "stops")
		if e != nil {
			return f, e
		}
		if lon < -180 || lon > 180 || lat < -90 || lat > 90 {
			return f, fmt.Errorf("%w: stop coordinate", ErrInvalidFeed)
		}
		stops[id] = true
		f.Stops = append(f.Stops, Stop{id, r["stop_name"], r["parent_station"], lon, lat})
	}
	for _, r := range t["routes.txt"] {
		id, e := need(r, "route_id", "routes")
		if e != nil {
			return f, e
		}
		if routes[id] {
			return f, fmt.Errorf("%w: duplicate route", ErrInvalidFeed)
		}
		if _, e = need(r, "route_type", "routes"); e != nil {
			return f, e
		}
		routes[id] = true
		f.Routes = append(f.Routes, Route{id, r["route_type"], r["route_short_name"], r["route_long_name"]})
	}
	for _, r := range t["calendar.txt"] {
		id, e := need(r, "service_id", "calendar")
		if e != nil {
			return f, e
		}
		start, e := need(r, "start_date", "calendar")
		if e != nil {
			return f, e
		}
		end, e := need(r, "end_date", "calendar")
		if e != nil {
			return f, e
		}
		if !gtfsDate(start) || !gtfsDate(end) || start > end {
			return f, fmt.Errorf("%w: calendar range", ErrInvalidFeed)
		}
		weekdays, err := calendarWeekdays(r)
		if err != nil {
			return f, err
		}
		services[id] = true
		f.Calendars = append(f.Calendars, Calendar{ServiceID: id, StartDate: start, EndDate: end, Weekdays: weekdays})
	}
	exceptions := map[string]struct{}{}
	for _, r := range t["calendar_dates.txt"] {
		id, e := need(r, "service_id", "calendar_dates")
		if e != nil {
			return f, e
		}
		date, e := need(r, "date", "calendar_dates")
		if e != nil || !gtfsDate(date) {
			return f, fmt.Errorf("%w: calendar_dates.date", ErrInvalidFeed)
		}
		typeValue, e := need(r, "exception_type", "calendar_dates")
		if e != nil || (typeValue != "1" && typeValue != "2") {
			return f, fmt.Errorf("%w: calendar_dates.exception_type", ErrInvalidFeed)
		}
		key := id + "\x00" + date
		if _, duplicate := exceptions[key]; duplicate {
			return f, fmt.Errorf("%w: duplicate calendar exception", ErrInvalidFeed)
		}
		exceptions[key] = struct{}{}
		services[id] = true
		f.Exceptions = append(f.Exceptions, CalendarException{ServiceID: id, Date: date, Added: typeValue == "1"})
	}
	for _, r := range t["trips.txt"] {
		id, e := need(r, "trip_id", "trips")
		if e != nil {
			return f, e
		}
		route, e := need(r, "route_id", "trips")
		if e != nil {
			return f, e
		}
		service, e := need(r, "service_id", "trips")
		if e != nil {
			return f, e
		}
		if trips[id] || !routes[route] || !services[service] {
			return f, fmt.Errorf("%w: trip reference", ErrInvalidFeed)
		}
		trips[id] = true
		f.Trips = append(f.Trips, Trip{id, route, service, r["shape_id"]})
	}
	if len(f.Stops) == 0 || len(f.Routes) == 0 || len(f.Trips) == 0 {
		return f, fmt.Errorf("%w: empty essential data", ErrInvalidFeed)
	}
	for _, r := range t["stop_times.txt"] {
		trip, e := need(r, "trip_id", "stop_times")
		if e != nil {
			return f, e
		}
		stop, e := need(r, "stop_id", "stop_times")
		if e != nil {
			return f, e
		}
		seq, e := integer(r, "stop_sequence", "stop_times")
		if e != nil {
			return f, e
		}
		if !trips[trip] || !stops[stop] {
			return f, fmt.Errorf("%w: stop time reference", ErrInvalidFeed)
		}
		a, ha, e := gtfsTime(r["arrival_time"])
		if e != nil {
			return f, e
		}
		d, hd, e := gtfsTime(r["departure_time"])
		if e != nil {
			return f, e
		}
		if ha && hd && d < a {
			return f, fmt.Errorf("%w: time order", ErrInvalidFeed)
		}
		f.StopTimes = append(f.StopTimes, StopTime{trip, stop, seq, a, d, ha, hd})
	}
	if len(f.StopTimes) == 0 {
		return f, fmt.Errorf("%w: empty stop_times", ErrInvalidFeed)
	}
	sort.Slice(f.StopTimes, func(i, j int) bool {
		return f.StopTimes[i].TripID+fmt.Sprintf("%09d", f.StopTimes[i].Sequence) < f.StopTimes[j].TripID+fmt.Sprintf("%09d", f.StopTimes[j].Sequence)
	})
	return f, nil
}
func gtfsDate(value string) bool {
	if len(value) != 8 {
		return false
	}
	parsed, err := time.Parse("20060102", value)
	return err == nil && parsed.Format("20060102") == value
}
func calendarWeekdays(row map[string]string) ([7]bool, error) {
	var days [7]bool
	for i, name := range []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"} {
		value, err := need(row, name, "calendar")
		if err != nil || (value != "0" && value != "1") {
			return days, fmt.Errorf("%w: calendar.%s", ErrInvalidFeed, name)
		}
		days[i] = value == "1"
	}
	return days, nil
}
func number(r map[string]string, k, t string) (float64, error) {
	v, e := need(r, k, t)
	if e != nil {
		return 0, e
	}
	n, e := strconv.ParseFloat(v, 64)
	if e != nil || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, fmt.Errorf("%w: %s.%s", ErrInvalidFeed, t, k)
	}
	return n, nil
}
func integer(r map[string]string, k, t string) (int, error) {
	v, e := need(r, k, t)
	if e != nil {
		return 0, e
	}
	n, e := strconv.Atoi(v)
	if e != nil || n < 0 {
		return 0, fmt.Errorf("%w: %s.%s", ErrInvalidFeed, t, k)
	}
	return n, nil
}
func gtfsTime(v string) (int, bool, error) {
	if v == "" {
		return 0, false, nil
	}
	p := strings.Split(v, ":")
	if len(p) != 3 {
		return 0, false, fmt.Errorf("%w: GTFS time", ErrInvalidFeed)
	}
	h, e := strconv.Atoi(p[0])
	if e != nil || h < 0 {
		return 0, false, ErrInvalidFeed
	}
	m, e := strconv.Atoi(p[1])
	if e != nil || m < 0 || m > 59 {
		return 0, false, ErrInvalidFeed
	}
	s, e := strconv.Atoi(p[2])
	if e != nil || s < 0 || s > 59 {
		return 0, false, ErrInvalidFeed
	}
	return h*3600 + m*60 + s, true, nil
}
