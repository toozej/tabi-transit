// Package notificationworker evaluates already-authorized notification
// deliveries. It deliberately has no Expo HTTP implementation: the runtime is
// a safe no-op until D-017 supplies reviewed credentials and an adapter.
package notificationworker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrPushGatewayUnavailable = errors.New("push gateway unavailable")
var ErrInvalidPushToken = errors.New("push token is invalid")

const maxAttempts = 3

type Delivery struct {
	ID, TokenID, SubscriptionID, Type, EntityID, DeepLink string
	PushToken                                             string
	ExpiresAt                                             time.Time
	Attempts                                              int
	QuietHours                                            *QuietHours
}

// QuietHours uses the installation's validated IANA timezone. Start equal to
// End means no quiet period; an overnight window is represented by Start > End.
type QuietHours struct {
	Start, End string
	TimeZone   string
}

type Store interface {
	Claim(context.Context, time.Time, int) ([]Delivery, error)
	MarkSent(context.Context, string, string, time.Time) error
	MarkRetry(context.Context, string, time.Time, string, time.Time) error
	MarkExpired(context.Context, string, time.Time) error
	MarkFailed(context.Context, string, string, time.Time) error
	DisableToken(context.Context, string, string, time.Time) error
}

// PushGateway is intentionally small so Expo Push can be added only after its
// external gate. Gateways must not log the supplied token or payload.
type PushGateway interface {
	Send(context.Context, Delivery) (ticketID string, err error)
}

// Receipt is a pre-sanitized provider outcome. It intentionally contains no
// provider response body, device token, credential, coordinate, or itinerary.
type Receipt struct {
	TicketID, Status, ErrorCode, SafeDetail string
	ReceivedAt                              time.Time
}

// ReceiptStore is separate from Store so delivery evaluation remains usable
// before a provider receipt adapter is approved.
type ReceiptStore interface {
	RecordReceipt(context.Context, Receipt, time.Time) error
}

type Service struct {
	Enabled bool
	Store   Store
	Gateway PushGateway
	Clock   func() time.Time
}

// RunOnce is a no-op while disabled. Once composed, it guarantees expired
// deliveries are terminal, honors quiet windows, and makes bounded retries.
func (s Service) RunOnce(ctx context.Context, limit int) (int, error) {
	if !s.Enabled {
		return 0, nil
	}
	if s.Store == nil || s.Gateway == nil {
		return 0, errors.New("notification worker is enabled without store or push gateway")
	}
	if limit < 1 || limit > 500 {
		return 0, errors.New("notification worker limit is outside bounds")
	}
	now := time.Now().UTC()
	if s.Clock != nil {
		now = s.Clock().UTC()
	}
	deliveries, err := s.Store.Claim(ctx, now, limit)
	if err != nil {
		return 0, fmt.Errorf("claim notification deliveries: %w", err)
	}
	for _, delivery := range deliveries {
		if !delivery.ExpiresAt.After(now) {
			if err := s.Store.MarkExpired(ctx, delivery.ID, now); err != nil {
				return 0, err
			}
			continue
		}
		quietUntil, quiet, err := quietUntil(delivery.QuietHours, now)
		if err != nil {
			if err := s.Store.MarkFailed(ctx, delivery.ID, "invalid_quiet_hours", now); err != nil {
				return 0, err
			}
			continue
		}
		if quiet {
			if !delivery.ExpiresAt.After(quietUntil) {
				if err := s.Store.MarkExpired(ctx, delivery.ID, now); err != nil {
					return 0, err
				}
			} else if err := s.Store.MarkRetry(ctx, delivery.ID, quietUntil, "quiet_hours", now); err != nil {
				return 0, err
			}
			continue
		}
		ticketID, sendErr := s.Gateway.Send(ctx, delivery)
		switch {
		case sendErr == nil:
			if err := s.Store.MarkSent(ctx, delivery.ID, ticketID, now); err != nil {
				return 0, err
			}
		case errors.Is(sendErr, ErrInvalidPushToken):
			if err := s.Store.DisableToken(ctx, delivery.TokenID, "invalid_token", now); err != nil {
				return 0, err
			}
			if err := s.Store.MarkFailed(ctx, delivery.ID, "invalid_token", now); err != nil {
				return 0, err
			}
		// Claim increments Attempts transactionally before delivery. A third
		// failed send is terminal; we do not create a fourth network attempt.
		case delivery.Attempts >= maxAttempts:
			if err := s.Store.MarkFailed(ctx, delivery.ID, "push_unavailable", now); err != nil {
				return 0, err
			}
		default:
			next := now.Add(time.Duration(delivery.Attempts+1) * time.Minute)
			if !delivery.ExpiresAt.After(next) {
				if err := s.Store.MarkExpired(ctx, delivery.ID, now); err != nil {
					return 0, err
				}
			} else if err := s.Store.MarkRetry(ctx, delivery.ID, next, "push_unavailable", now); err != nil {
				return 0, err
			}
		}
	}
	return len(deliveries), nil
}

// ProcessReceipt is deliberately inert while delivery is disabled. A future
// provider adapter may pass only an already-classified Receipt into this
// boundary after D-017 and D-018 are approved.
func (s Service) ProcessReceipt(ctx context.Context, receipt Receipt) error {
	if !s.Enabled {
		return nil
	}
	store, ok := s.Store.(ReceiptStore)
	if !ok {
		return errors.New("notification worker is enabled without receipt store")
	}
	now := time.Now().UTC()
	if s.Clock != nil {
		now = s.Clock().UTC()
	}
	return store.RecordReceipt(ctx, receipt, now)
}

func quietUntil(window *QuietHours, now time.Time) (time.Time, bool, error) {
	if window == nil || window.Start == window.End {
		return time.Time{}, false, nil
	}
	location, err := time.LoadLocation(window.TimeZone)
	if err != nil {
		return time.Time{}, false, err
	}
	start, err := time.Parse("15:04", window.Start)
	if err != nil {
		return time.Time{}, false, err
	}
	end, err := time.Parse("15:04", window.End)
	if err != nil {
		return time.Time{}, false, err
	}
	local := now.In(location)
	minute := local.Hour()*60 + local.Minute()
	startMinute, endMinute := start.Hour()*60+start.Minute(), end.Hour()*60+end.Minute()
	inWindow := (startMinute < endMinute && minute >= startMinute && minute < endMinute) || (startMinute > endMinute && (minute >= startMinute || minute < endMinute))
	if !inWindow {
		return time.Time{}, false, nil
	}
	endDay := local
	if startMinute > endMinute && minute >= startMinute {
		endDay = endDay.AddDate(0, 0, 1)
	}
	return time.Date(endDay.Year(), endDay.Month(), endDay.Day(), end.Hour(), end.Minute(), 0, 0, location).UTC(), true, nil
}
