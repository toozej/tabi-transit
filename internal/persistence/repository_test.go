package persistence

import (
	"testing"
	"time"
)

func TestValidatePublicID(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"trimet:vehicle:2901", "fixture:vehicle:vehicle:1"} {
		if err := ValidatePublicID(id, "vehicle"); err != nil {
			t.Fatalf("%q rejected: %v", id, err)
		}
	}
	for _, id := range []string{"2901", "trimet:route:20", "trimet:vehicle:\n2901"} {
		if err := ValidatePublicID(id, "vehicle"); err == nil {
			t.Fatalf("%q accepted", id)
		}
	}
}

func TestServiceSecondsAfterMidnight(t *testing.T) {
	t.Parallel()
	seconds, err := ServiceSeconds(25, 15, 30)
	if err != nil || seconds != 90930 {
		t.Fatalf("got %d, %v", seconds, err)
	}
}

func TestServiceInstantUsesAgencyTimezoneAndPreservesAfterMidnight(t *testing.T) {
	t.Parallel()
	zone, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2026, 7, 22, 0, 0, 0, 0, zone)
	got := serviceInstant(date, 25*3600+15*60, zone)
	want := time.Date(2026, 7, 23, 8, 15, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %s want %s", got, want)
	}
}

func FuzzValidatePublicID(f *testing.F) {
	for _, seed := range []string{"trimet:vehicle:2901", "", "trimet:route:20", "x:vehicle:y"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, id string) {
		err := ValidatePublicID(id, "vehicle")
		if err == nil && (len(id) == 0 || len(id) > 512) {
			t.Fatalf("invalid length accepted: %d", len(id))
		}
	})
}
