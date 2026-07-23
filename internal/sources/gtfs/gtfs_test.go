package gtfs

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func archive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for n, v := range files {
		f, e := w.Create(n)
		if e != nil {
			t.Fatal(e)
		}
		_, _ = f.Write([]byte(v))
	}
	if e := w.Close(); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func valid() map[string]string {
	return map[string]string{"stops.txt": "stop_id,stop_name,stop_lon,stop_lat\ns1,One,-122.6,45.5\n", "routes.txt": "route_id,route_type\nr1,3\n", "trips.txt": "route_id,service_id,trip_id\nr1,weekday,t1\n", "calendar.txt": "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nweekday,1,1,1,1,1,0,0,20260101,20261231\n", "calendar_dates.txt": "service_id,date,exception_type\nweekday,20260704,2\nweekday,20260705,1\n", "stop_times.txt": "trip_id,arrival_time,departure_time,stop_id,stop_sequence\nt1,25:15:30,25:30:00,s1,1\n"}
}
func TestReadAfterMidnight(t *testing.T) {
	f, e := Read(bytes.NewReader(archive(t, valid())), DefaultArchivePolicy())
	if e != nil {
		t.Fatal(e)
	}
	if f.StopTimes[0].ArrivalSeconds != 90930 {
		t.Fatalf("got %d", f.StopTimes[0].ArrivalSeconds)
	}
	if !f.Calendars[0].Weekdays[0] || f.Calendars[0].Weekdays[5] || len(f.Exceptions) != 2 || f.Exceptions[0].Added {
		t.Fatalf("calendar data not retained: %#v %#v", f.Calendars, f.Exceptions)
	}
}
func TestReadRejectsInvalidCalendarSemantics(t *testing.T) {
	for name, mutate := range map[string]func(map[string]string){
		"bad_weekday": func(m map[string]string) {
			m["calendar.txt"] = "service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date\nweekday,2,1,1,1,1,0,0,20260101,20261231\n"
		},
		"bad_exception": func(m map[string]string) {
			m["calendar_dates.txt"] = "service_id,date,exception_type\nweekday,20260230,1\n"
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Read(bytes.NewReader(archive(t, func() map[string]string { m := valid(); mutate(m); return m }())), DefaultArchivePolicy())
			if err == nil {
				t.Fatal("accepted invalid calendar")
			}
		})
	}
}

func TestReadRetainsValidatedAgencyTimezone(t *testing.T) {
	t.Parallel()
	files := valid()
	files["agency.txt"] = "agency_id,agency_name,agency_timezone\na,Fixture Transit,America/Los_Angeles\n"
	f, err := Read(bytes.NewReader(archive(t, files)), DefaultArchivePolicy())
	if err != nil || len(f.AgencyTimezones) != 1 || f.AgencyTimezones[0].Timezone != "America/Los_Angeles" {
		t.Fatalf("timezone=%#v err=%v", f.AgencyTimezones, err)
	}
}

func TestReadRejectsInvalidOrAmbiguousAgencyTimezone(t *testing.T) {
	t.Parallel()
	for _, agency := range []string{
		"agency_id,agency_name,agency_timezone\na,Fixture,Not/AZone\n",
		"agency_id,agency_name,agency_timezone\na,Fixture,America/Los_Angeles\nb,Other,America/New_York\n",
	} {
		files := valid()
		files["agency.txt"] = agency
		if _, err := Read(bytes.NewReader(archive(t, files)), DefaultArchivePolicy()); err == nil {
			t.Fatal("accepted unsafe agency timezone")
		}
	}
}
func TestReadRejectsMaliciousAndInvalid(t *testing.T) {
	for n, mut := range map[string]func(map[string]string){"traversal": func(m map[string]string) { m["../stops.txt"] = m["stops.txt"]; delete(m, "stops.txt") }, "bad-reference": func(m map[string]string) { m["trips.txt"] = "route_id,service_id,trip_id\nmissing,weekday,t1\n" }, "bad-csv": func(m map[string]string) {
		m["stops.txt"] = "stop_id,stop_name,stop_lon,stop_lat\ns1,unterminated,-1,2\n\""
	}, "empty-essential": func(m map[string]string) { m["routes.txt"] = "route_id,route_type\n" }, "no-service": func(m map[string]string) { delete(m, "calendar.txt"); delete(m, "calendar_dates.txt") }} {
		t.Run(n, func(t *testing.T) {
			m := valid()
			mut(m)
			_, e := Read(bytes.NewReader(archive(t, m)), DefaultArchivePolicy())
			if e == nil {
				t.Fatal("accepted invalid feed")
			}
		})
	}
}
func TestReadRejectsCompressionPolicy(t *testing.T) {
	m := valid()
	m["extra.txt"] = strings.Repeat("x", 8192)
	p := DefaultArchivePolicy()
	p.MaxCompressionRatio = 1
	_, e := Read(bytes.NewReader(archive(t, m)), p)
	if e == nil {
		t.Fatal("accepted compressed archive")
	}
}
