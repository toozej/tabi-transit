package transitdataspike

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// StaticFeed is the deliberately small subset used to prove CSV boundary
// handling before the production importer is designed.
type StaticFeed struct {
	Stops     []Stop
	Routes    []Route
	Trips     []Trip
	StopTimes []StopTime
}

type Stop struct {
	ID, Name            string
	Latitude, Longitude float64
}
type Route struct{ ID, Type string }
type Trip struct{ ID, RouteID, ServiceID string }
type StopTime struct {
	TripID, Arrival, Departure string
	StopSequence               int
}

// ParseStaticGTFS parses required GTFS CSV files from the supplied readers.
// It intentionally does not fetch archives or implement production validation.
func ParseStaticGTFS(files map[string]io.Reader) (StaticFeed, error) {
	required := []string{"stops.txt", "routes.txt", "trips.txt", "stop_times.txt"}
	for _, name := range required {
		if files[name] == nil {
			return StaticFeed{}, fmt.Errorf("required GTFS file missing: %s", name)
		}
	}
	stopsRows, err := csvRows(files["stops.txt"])
	if err != nil {
		return StaticFeed{}, fmt.Errorf("stops.txt: %w", err)
	}
	routeRows, err := csvRows(files["routes.txt"])
	if err != nil {
		return StaticFeed{}, fmt.Errorf("routes.txt: %w", err)
	}
	tripRows, err := csvRows(files["trips.txt"])
	if err != nil {
		return StaticFeed{}, fmt.Errorf("trips.txt: %w", err)
	}
	stopTimeRows, err := csvRows(files["stop_times.txt"])
	if err != nil {
		return StaticFeed{}, fmt.Errorf("stop_times.txt: %w", err)
	}

	feed := StaticFeed{}
	stopIDs, routeIDs, tripIDs := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, row := range stopsRows {
		id := row["stop_id"]
		if id == "" || stopIDs[id] {
			return StaticFeed{}, fmt.Errorf("invalid or duplicate stop_id %q", id)
		}
		stopIDs[id] = true
		lat, err := strconv.ParseFloat(row["stop_lat"], 64)
		if err != nil || lat < -90 || lat > 90 {
			return StaticFeed{}, fmt.Errorf("stop %s has invalid latitude", id)
		}
		lon, err := strconv.ParseFloat(row["stop_lon"], 64)
		if err != nil || lon < -180 || lon > 180 {
			return StaticFeed{}, fmt.Errorf("stop %s has invalid longitude", id)
		}
		feed.Stops = append(feed.Stops, Stop{ID: id, Name: row["stop_name"], Latitude: lat, Longitude: lon})
	}
	for _, row := range routeRows {
		id := row["route_id"]
		if id == "" || routeIDs[id] {
			return StaticFeed{}, fmt.Errorf("invalid or duplicate route_id %q", id)
		}
		routeIDs[id] = true
		feed.Routes = append(feed.Routes, Route{ID: id, Type: row["route_type"]})
	}
	for _, row := range tripRows {
		id, routeID := row["trip_id"], row["route_id"]
		if id == "" || tripIDs[id] || !routeIDs[routeID] {
			return StaticFeed{}, fmt.Errorf("invalid trip %q or route reference %q", id, routeID)
		}
		tripIDs[id] = true
		feed.Trips = append(feed.Trips, Trip{ID: id, RouteID: routeID, ServiceID: row["service_id"]})
	}
	for _, row := range stopTimeRows {
		tripID := row["trip_id"]
		if !tripIDs[tripID] || !stopIDs[row["stop_id"]] {
			return StaticFeed{}, fmt.Errorf("stop time has missing trip or stop reference")
		}
		arrival, err := ServiceDaySeconds(row["arrival_time"])
		if err != nil {
			return StaticFeed{}, fmt.Errorf("arrival_time: %w", err)
		}
		departure, err := ServiceDaySeconds(row["departure_time"])
		if err != nil {
			return StaticFeed{}, fmt.Errorf("departure_time: %w", err)
		}
		if departure < arrival {
			return StaticFeed{}, fmt.Errorf("stop time departure precedes arrival")
		}
		sequence, err := strconv.Atoi(row["stop_sequence"])
		if err != nil || sequence < 1 {
			return StaticFeed{}, fmt.Errorf("invalid stop_sequence")
		}
		feed.StopTimes = append(feed.StopTimes, StopTime{TripID: tripID, Arrival: row["arrival_time"], Departure: row["departure_time"], StopSequence: sequence})
	}
	return feed, nil
}

// ServiceDaySeconds retains GTFS times beyond 24:00 in service-day semantics.
func ServiceDaySeconds(value string) (int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid GTFS time %q", value)
	}
	hour, e1 := strconv.Atoi(parts[0])
	minute, e2 := strconv.Atoi(parts[1])
	second, e3 := strconv.Atoi(parts[2])
	if e1 != nil || e2 != nil || e3 != nil || hour < 0 || minute < 0 || minute > 59 || second < 0 || second > 59 {
		return 0, fmt.Errorf("invalid GTFS time %q", value)
	}
	return hour*3600 + minute*60 + second, nil
}

func csvRows(reader io.Reader) ([]map[string]string, error) {
	r := csv.NewReader(reader)
	r.FieldsPerRecord = -1
	headers, err := r.Read()
	if err != nil {
		return nil, err
	}
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}
	var rows []map[string]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) != len(headers) {
			return nil, fmt.Errorf("expected %d columns, got %d", len(headers), len(record))
		}
		row := map[string]string{}
		for i, header := range headers {
			row[header] = strings.TrimSpace(record[i])
		}
		rows = append(rows, row)
	}
	return rows, nil
}
