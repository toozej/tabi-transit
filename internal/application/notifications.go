package application

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	FeatureNotifications          = "notifications"
	ReasonNotificationGatePending = "notification_delivery_and_persistence_pending"
)

// NotificationStore is the persistence boundary for installation ownership and
// subscriptions. A future implementation stores only a verifier for the opaque
// installation credential; the raw credential and push token never belong in
// logs or application errors.
type NotificationStore interface {
	CreateInstallation(context.Context, InstallationRegistration) (Installation, error)
	RegisterPushToken(context.Context, InstallationCredential, PushTokenRegistration) error
	DeleteInstallation(context.Context, InstallationCredential, string) error
	ListSubscriptions(context.Context, InstallationCredential) ([]Subscription, error)
	CreateSubscription(context.Context, InstallationCredential, SubscriptionDraft) (Subscription, error)
	DeleteSubscription(context.Context, InstallationCredential, string) error
}

type NotificationFeatures struct{ Store NotificationStore }

type InstallationRegistration struct {
	Platform, Locale, TimeZone, AppVersion string
}
type Installation struct {
	ID string
	// Credential is a one-time opaque secret. It is only returned by a future
	// enabled registration handler over TLS and must never be logged or stored
	// in a fixture.
	Credential string
}
type InstallationCredential string

type PushTokenRegistration struct {
	Token, Platform string
}

type SubscriptionType string

const (
	SubscriptionServiceAlert      SubscriptionType = "service_alert"
	SubscriptionDepartureReminder SubscriptionType = "departure_reminder"
)

type QuietHours struct {
	Start, End string
	TimeZone   string
}
type SubscriptionDraft struct {
	Type                                SubscriptionType
	RouteIDs, StopIDs, Modes, SourceIDs []string
	TripID                              string
	RemindAt, ExpiresAt                 *time.Time
	QuietHours                          *QuietHours
}
type Subscription struct {
	ID                                  string
	Type                                SubscriptionType
	RouteIDs, StopIDs, Modes, SourceIDs []string
	TripID                              string
	RemindAt, ExpiresAt                 *time.Time
	QuietHours                          *QuietHours
}

func (s Service) CreateInstallation(ctx context.Context, input InstallationRegistration) (Installation, error) {
	if err := validateInstallationRegistration(input); err != nil {
		return Installation{}, err
	}
	if s.Notifications.Store == nil {
		return Installation{}, disabledNotification()
	}
	return s.Notifications.Store.CreateInstallation(ctx, input)
}
func (s Service) RegisterPushToken(ctx context.Context, credential InstallationCredential, input PushTokenRegistration) error {
	if err := validateCredential(credential); err != nil {
		return err
	}
	if err := validatePushToken(input); err != nil {
		return err
	}
	if s.Notifications.Store == nil {
		return disabledNotification()
	}
	return s.Notifications.Store.RegisterPushToken(ctx, credential, input)
}
func (s Service) DeleteInstallation(ctx context.Context, credential InstallationCredential, id string) error {
	if err := validateCredential(credential); err != nil {
		return err
	}
	if invalidOpaqueID(id) {
		return errors.New("invalid installation ID")
	}
	if s.Notifications.Store == nil {
		return disabledNotification()
	}
	return s.Notifications.Store.DeleteInstallation(ctx, credential, id)
}
func (s Service) ListSubscriptions(ctx context.Context, credential InstallationCredential) ([]Subscription, error) {
	if err := validateCredential(credential); err != nil {
		return nil, err
	}
	if s.Notifications.Store == nil {
		return nil, disabledNotification()
	}
	return s.Notifications.Store.ListSubscriptions(ctx, credential)
}
func (s Service) CreateSubscription(ctx context.Context, credential InstallationCredential, input SubscriptionDraft) (Subscription, error) {
	if err := validateCredential(credential); err != nil {
		return Subscription{}, err
	}
	if err := validateSubscription(input, s.now()); err != nil {
		return Subscription{}, err
	}
	if s.Notifications.Store == nil {
		return Subscription{}, disabledNotification()
	}
	return s.Notifications.Store.CreateSubscription(ctx, credential, input)
}
func (s Service) DeleteSubscription(ctx context.Context, credential InstallationCredential, id string) error {
	if err := validateCredential(credential); err != nil {
		return err
	}
	if invalidOpaqueID(id) {
		return errors.New("invalid subscription ID")
	}
	if s.Notifications.Store == nil {
		return disabledNotification()
	}
	return s.Notifications.Store.DeleteSubscription(ctx, credential, id)
}

func disabledNotification() error {
	return &FeatureDisabledError{Feature: FeatureNotifications, Reason: ReasonNotificationGatePending}
}
func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func validateInstallationRegistration(input InstallationRegistration) error {
	if !oneOf(input.Platform, "ios", "android") || invalidShort(input.Locale, 35) || invalidShort(input.AppVersion, 100) || !validZone(input.TimeZone) {
		return errors.New("invalid installation registration")
	}
	return nil
}
func validatePushToken(input PushTokenRegistration) error {
	// Deliberately only validate shape. The token itself is never surfaced from
	// this package, including in errors.
	if !oneOf(input.Platform, "ios", "android") || invalidShort(input.Token, 4096) || len(strings.TrimSpace(input.Token)) < 8 {
		return errors.New("invalid push token registration")
	}
	return nil
}
func validateCredential(value InstallationCredential) error {
	if invalidShort(string(value), 1024) || len(strings.TrimSpace(string(value))) < 32 {
		return errors.New("invalid installation credential")
	}
	return nil
}
func validateSubscription(input SubscriptionDraft, now time.Time) error {
	if input.Type != SubscriptionServiceAlert && input.Type != SubscriptionDepartureReminder {
		return errors.New("invalid subscription")
	}
	if input.ExpiresAt == nil || !input.ExpiresAt.After(now) || input.ExpiresAt.After(now.Add(90*24*time.Hour)) {
		return errors.New("invalid subscription expiry")
	}
	if input.QuietHours != nil && (!validClock(input.QuietHours.Start) || !validClock(input.QuietHours.End) || !validZone(input.QuietHours.TimeZone)) {
		return errors.New("invalid quiet hours")
	}
	for _, mode := range input.Modes {
		if !oneOf(mode, "bus", "light_rail", "commuter_rail", "streetcar", "aerial_tram", "unknown") {
			return errors.New("invalid subscription scope")
		}
	}
	for _, id := range append(append([]string{}, input.RouteIDs...), input.StopIDs...) {
		if invalidOpaqueID(id) {
			return errors.New("invalid subscription scope")
		}
	}
	for _, sourceID := range input.SourceIDs {
		if invalidShort(sourceID, 128) {
			return errors.New("invalid subscription scope")
		}
	}
	if input.Type == SubscriptionServiceAlert && len(input.RouteIDs)+len(input.StopIDs)+len(input.Modes)+len(input.SourceIDs) == 0 {
		return errors.New("invalid subscription scope")
	}
	if input.Type == SubscriptionDepartureReminder && (invalidOpaqueID(input.TripID) || input.RemindAt == nil || !input.RemindAt.After(now) || !input.RemindAt.Before(*input.ExpiresAt)) {
		return errors.New("invalid departure reminder")
	}
	return nil
}
func validZone(value string) bool {
	if invalidShort(value, 100) {
		return false
	}
	_, err := time.LoadLocation(value)
	return err == nil
}
func validClock(value string) bool {
	_, err := time.Parse("15:04", value)
	return err == nil && len(value) == 5
}
func invalidShort(value string, max int) bool {
	return strings.TrimSpace(value) == "" || len(value) > max || containsControl(value)
}
func invalidOpaqueID(value string) bool { return invalidShort(value, 255) }
func oneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}
