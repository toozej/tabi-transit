package persistence

import "testing"

func TestClassifyAdherenceUsesTripUpdateEvidence(t *testing.T) {
	early, onTime, late := int32(-61), int32(60), int32(61)
	for _, tc := range []struct {
		name, trip, stop   string
		arrival, departure *int32
		want               AdherenceStatus
	}{
		{"unknown", "", "", nil, nil, AdherenceUnknown},
		{"canceled", "CANCELED", "", &late, nil, AdherenceCanceled},
		{"skipped", "", "SKIPPED", &late, nil, AdherenceCanceled},
		{"early", "", "", &early, nil, AdherenceEarly},
		{"on time boundary", "", "", &onTime, nil, AdherenceOnTime},
		{"late", "", "", &late, nil, AdherenceLate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyAdherence(tc.trip, tc.stop, tc.arrival, tc.departure); got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}
