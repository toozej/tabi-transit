package persistence

import (
	"testing"
	"time"
)

func TestVehicleHistoryRetentionIsThirtyDays(t *testing.T) {
	if vehicleHistoryRetention != 30*24*time.Hour {
		t.Fatalf("vehicle history retention = %s, want 30 days", vehicleHistoryRetention)
	}
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	if got, want := now.Add(-vehicleHistoryRetention), time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("retention cutoff = %s, want %s", got, want)
	}
}
