package notificationworker

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	deliveries []Delivery
	calls      []string
}

func (s *fakeStore) Claim(_ context.Context, _ time.Time, _ int) ([]Delivery, error) {
	return s.deliveries, nil
}
func (s *fakeStore) MarkSent(_ context.Context, id, _ string, _ time.Time) error {
	s.calls = append(s.calls, "sent:"+id)
	return nil
}
func (s *fakeStore) MarkRetry(_ context.Context, id string, _ time.Time, code string, _ time.Time) error {
	s.calls = append(s.calls, "retry:"+id+":"+code)
	return nil
}
func (s *fakeStore) MarkExpired(_ context.Context, id string, _ time.Time) error {
	s.calls = append(s.calls, "expired:"+id)
	return nil
}
func (s *fakeStore) MarkFailed(_ context.Context, id, code string, _ time.Time) error {
	s.calls = append(s.calls, "failed:"+id+":"+code)
	return nil
}
func (s *fakeStore) DisableToken(_ context.Context, id, code string, _ time.Time) error {
	s.calls = append(s.calls, "disabled:"+id+":"+code)
	return nil
}

type fakeGateway struct{ err error }

func (g fakeGateway) Send(_ context.Context, _ Delivery) (string, error) { return "ticket-safe", g.err }

func TestRunOnceIsNoopWhenDisabled(t *testing.T) {
	t.Parallel()
	count, err := (Service{}).RunOnce(context.Background(), 1)
	if count != 0 || err != nil {
		t.Fatalf("disabled result = %d, %v", count, err)
	}
}
func TestRunOnceNeverRetriesExpiredAndDisablesInvalidTokens(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{deliveries: []Delivery{{ID: "expired", ExpiresAt: now}, {ID: "bad-token", TokenID: "token-1", ExpiresAt: now.Add(time.Hour)}}}
	service := Service{Enabled: true, Store: store, Gateway: fakeGateway{err: ErrInvalidPushToken}, Clock: func() time.Time { return now }}
	if _, err := service.RunOnce(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	want := []string{"expired:expired", "disabled:token-1:invalid_token", "failed:bad-token:invalid_token"}
	if len(store.calls) != len(want) {
		t.Fatalf("calls = %#v", store.calls)
	}
	for i := range want {
		if store.calls[i] != want[i] {
			t.Fatalf("calls = %#v", store.calls)
		}
	}
}
func TestRunOnceDefersQuietHoursOnlyBeforeExpiry(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 23, 6, 30, 0, 0, time.UTC) // 23:30 PDT previous day
	store := &fakeStore{deliveries: []Delivery{{ID: "defer", ExpiresAt: now.Add(12 * time.Hour), QuietHours: &QuietHours{Start: "22:00", End: "07:00", TimeZone: "America/Los_Angeles"}}, {ID: "expire", ExpiresAt: now.Add(time.Hour), QuietHours: &QuietHours{Start: "22:00", End: "07:00", TimeZone: "America/Los_Angeles"}}}}
	service := Service{Enabled: true, Store: store, Gateway: fakeGateway{err: errors.New("must not send")}, Clock: func() time.Time { return now }}
	if _, err := service.RunOnce(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	want := []string{"retry:defer:quiet_hours", "expired:expire"}
	for i := range want {
		if store.calls[i] != want[i] {
			t.Fatalf("calls = %#v", store.calls)
		}
	}
}
func TestRunOnceBoundsRetries(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	store := &fakeStore{deliveries: []Delivery{{ID: "retry", ExpiresAt: now.Add(time.Hour), Attempts: 1}, {ID: "max", ExpiresAt: now.Add(time.Hour), Attempts: 3}}}
	service := Service{Enabled: true, Store: store, Gateway: fakeGateway{err: ErrPushGatewayUnavailable}, Clock: func() time.Time { return now }}
	if _, err := service.RunOnce(context.Background(), 2); err != nil {
		t.Fatal(err)
	}
	want := []string{"retry:retry:push_unavailable", "failed:max:push_unavailable"}
	for i := range want {
		if store.calls[i] != want[i] {
			t.Fatalf("calls = %#v", store.calls)
		}
	}
}
