package persistence

import "strings"

// AdherenceStatus is evidence-based; GPS observations alone remain unknown.
type AdherenceStatus string

const (
	AdherenceUnknown  AdherenceStatus = "unknown"
	AdherenceCanceled AdherenceStatus = "canceled"
	AdherenceEarly    AdherenceStatus = "early"
	AdherenceOnTime   AdherenceStatus = "on_time"
	AdherenceLate     AdherenceStatus = "late"
)

const adherenceOnTimeThresholdSeconds int32 = 60

func ClassifyAdherence(tripRelationship, stopRelationship string, arrivalDelay, departureDelay *int32) AdherenceStatus {
	if strings.EqualFold(tripRelationship, "CANCELED") || strings.EqualFold(stopRelationship, "SKIPPED") || strings.EqualFold(stopRelationship, "CANCELED") {
		return AdherenceCanceled
	}
	delay := arrivalDelay
	if delay == nil {
		delay = departureDelay
	}
	if delay == nil {
		return AdherenceUnknown
	}
	if *delay < -adherenceOnTimeThresholdSeconds {
		return AdherenceEarly
	}
	if *delay > adherenceOnTimeThresholdSeconds {
		return AdherenceLate
	}
	return AdherenceOnTime
}
