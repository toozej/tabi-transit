package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

func notificationDraft(now time.Time) SubscriptionDraft {
	expires := now.Add(24 * time.Hour)
	return SubscriptionDraft{Type: SubscriptionServiceAlert, RouteIDs: []string{"trimet:route:20"}, ExpiresAt: &expires, QuietHours: &QuietHours{Start: "22:00", End: "07:00", TimeZone: "America/Los_Angeles"}}
}

func TestNotificationsFailClosedWithoutPersistence(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	service := Service{Now: func() time.Time { return now }}
	_, err := service.CreateInstallation(context.Background(), InstallationRegistration{Platform: "ios", Locale: "en-US", TimeZone: "America/Los_Angeles", AppVersion: "1.0.0"})
	if !errors.Is(err, ErrFeatureDisabled) {
		t.Fatalf("registration error = %v", err)
	}
	credential := InstallationCredential("0123456789abcdef0123456789abcdef")
	_, err = service.CreateSubscription(context.Background(), credential, notificationDraft(now))
	var disabledErr *FeatureDisabledError
	if !errors.As(err, &disabledErr) || disabledErr.Feature != FeatureNotifications || disabledErr.Reason != ReasonNotificationGatePending {
		t.Fatalf("subscription error = %v", err)
	}
}

func TestNotificationValidationRejectsUnsafeOrLateInputs(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	service := Service{Now: func() time.Time { return now }}
	if _, err := service.CreateInstallation(context.Background(), InstallationRegistration{Platform: "ios", Locale: "en-US", TimeZone: "Not/AZone", AppVersion: "1.0"}); err == nil {
		t.Fatal("invalid zone accepted")
	}
	if err := service.RegisterPushToken(context.Background(), InstallationCredential("short"), PushTokenRegistration{Platform: "ios", Token: "ExponentPushToken[private]"}); err == nil {
		t.Fatal("short credential accepted")
	}
	bad := notificationDraft(now)
	past := now.Add(-time.Minute)
	bad.ExpiresAt = &past
	if _, err := service.CreateSubscription(context.Background(), InstallationCredential("0123456789abcdef0123456789abcdef"), bad); err == nil {
		t.Fatal("past expiry accepted")
	}
	bad = notificationDraft(now)
	bad.QuietHours.TimeZone = "invalid\nzone"
	if _, err := service.CreateSubscription(context.Background(), InstallationCredential("0123456789abcdef0123456789abcdef"), bad); err == nil {
		t.Fatal("unsafe quiet hours accepted")
	}
}
