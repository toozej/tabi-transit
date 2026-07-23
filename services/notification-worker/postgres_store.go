package notificationworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/toozej/tabi-transit/internal/persistence"
	"github.com/toozej/tabi-transit/internal/persistence/sqlcgen"
)

const defaultClaimLease = 2 * time.Minute

var ErrNotificationStateConflict = errors.New("notification delivery state changed")
var ErrInvalidNotificationReceipt = errors.New("invalid notification receipt")
var errUnsafeDeliveryPayload = errors.New("claimed delivery has unsafe payload")
var errTokenDecryption = errors.New("cannot decrypt claimed push token")

// PostgresStore is the notification-worker persistence boundary. It receives
// ciphertext only from the database and decrypts it only after the delivery has
// been atomically claimed. It never logs token ciphertext or plaintext.
type PostgresStore struct {
	queries      notificationQueries
	beginReceipt func(context.Context) (receiptTransaction, error)
	protector    persistence.PushTokenProtector
	claimLease   time.Duration
}

type notificationQueries interface {
	ClaimNotificationDeliveries(context.Context, sqlcgen.ClaimNotificationDeliveriesParams) ([]sqlcgen.ClaimNotificationDeliveriesRow, error)
	DisablePushToken(context.Context, sqlcgen.DisablePushTokenParams) error
	MarkNotificationDeliverySent(context.Context, sqlcgen.MarkNotificationDeliverySentParams) (int64, error)
	MarkNotificationDeliveryRetry(context.Context, sqlcgen.MarkNotificationDeliveryRetryParams) (int64, error)
	MarkNotificationDeliveryExpired(context.Context, sqlcgen.MarkNotificationDeliveryExpiredParams) (int64, error)
	MarkNotificationDeliveryFailed(context.Context, sqlcgen.MarkNotificationDeliveryFailedParams) (int64, error)
}

type receiptQueries interface {
	RecordNotificationReceipt(context.Context, sqlcgen.RecordNotificationReceiptParams) (string, error)
	DisableTokenForInvalidReceipt(context.Context, sqlcgen.DisableTokenForInvalidReceiptParams) (int64, error)
}

// receiptTransaction is a deliberately small fakeable seam for the only
// multi-statement notification state transition.
type receiptTransaction interface {
	receiptQueries
	Commit(context.Context) error
	Rollback(context.Context) error
}

type pgxReceiptTransaction struct {
	tx pgx.Tx
	q  *sqlcgen.Queries
}

func (t pgxReceiptTransaction) RecordNotificationReceipt(ctx context.Context, arg sqlcgen.RecordNotificationReceiptParams) (string, error) {
	return t.q.RecordNotificationReceipt(ctx, arg)
}

func (t pgxReceiptTransaction) DisableTokenForInvalidReceipt(ctx context.Context, arg sqlcgen.DisableTokenForInvalidReceiptParams) (int64, error) {
	return t.q.DisableTokenForInvalidReceipt(ctx, arg)
}

func (t pgxReceiptTransaction) Commit(ctx context.Context) error   { return t.tx.Commit(ctx) }
func (t pgxReceiptTransaction) Rollback(ctx context.Context) error { return t.tx.Rollback(ctx) }

// NewPostgresStore composes a worker store without exposing a token plaintext
// outside the worker. A nil pool or protector is rejected before the worker can
// be enabled.
func NewPostgresStore(pool *pgxpool.Pool, protector persistence.PushTokenProtector) (*PostgresStore, error) {
	if pool == nil {
		return nil, errors.New("notification store requires database pool")
	}
	if protector == nil {
		return nil, errors.New("notification store requires push token protector")
	}
	return &PostgresStore{
		queries:    sqlcgen.New(pool),
		protector:  protector,
		claimLease: defaultClaimLease,
		beginReceipt: func(ctx context.Context) (receiptTransaction, error) {
			tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				return nil, err
			}
			return pgxReceiptTransaction{tx: tx, q: sqlcgen.New(tx)}, nil
		},
	}, nil
}

func (s *PostgresStore) Claim(ctx context.Context, now time.Time, limit int) ([]Delivery, error) {
	if s == nil || s.queries == nil || s.protector == nil {
		return nil, errors.New("notification store is not configured")
	}
	if limit < 1 || limit > 500 {
		return nil, errors.New("notification claim limit is outside bounds")
	}
	lease := s.claimLease
	if lease <= 0 {
		lease = defaultClaimLease
	}
	now = now.UTC()
	rows, err := s.queries.ClaimNotificationDeliveries(ctx, sqlcgen.ClaimNotificationDeliveriesParams{
		NowAt: pgTimestamp(now), MaxAttempts: maxAttempts, RowLimit: int32(limit), ClaimUntil: pgTimestamp(now.Add(lease)),
	})
	if err != nil {
		return nil, fmt.Errorf("claim notification deliveries: %w", err)
	}
	deliveries := make([]Delivery, 0, len(rows))
	for _, row := range rows {
		delivery, err := s.deliveryFromClaim(row)
		if err != nil {
			// A malformed encrypted row cannot safely be sent. It has already
			// been leased, so make it terminal instead of leaving a retry loop.
			if id, idErr := uuidString(row.ID); idErr == nil {
				parsedID, parseErr := parseUUID(id)
				if parseErr != nil {
					return nil, parseErr
				}
				code := "invalid_delivery_payload"
				if errors.Is(err, errTokenDecryption) {
					code = "token_decryption_failed"
				}
				_, markErr := s.queries.MarkNotificationDeliveryFailed(ctx, sqlcgen.MarkNotificationDeliveryFailedParams{
					ID: parsedID, ErrorCode: pgText(code), UpdatedAt: pgTimestamp(now),
				})
				if markErr != nil {
					return nil, fmt.Errorf("mark malformed notification delivery: %w", markErr)
				}
			}
			continue
		}
		deliveries = append(deliveries, delivery)
	}
	return deliveries, nil
}

func (s *PostgresStore) deliveryFromClaim(row sqlcgen.ClaimNotificationDeliveriesRow) (Delivery, error) {
	if !row.ExpiresAt.Valid {
		return Delivery{}, errors.New("claimed delivery has no expiry")
	}
	id, err := uuidString(row.ID)
	if err != nil {
		return Delivery{}, err
	}
	tokenID, err := uuidString(row.PushTokenID)
	if err != nil {
		return Delivery{}, err
	}
	subscriptionID, err := uuidString(row.SubscriptionID)
	if err != nil {
		return Delivery{}, err
	}
	var payload struct {
		EntityID       string `json:"entityId"`
		DeepLink       string `json:"deepLink"`
		SubscriptionID string `json:"subscriptionId"`
	}
	if len(row.Payload) == 0 || len(row.Payload) > 4096 || json.Unmarshal(row.Payload, &payload) != nil || payload.EntityID == "" || payload.DeepLink == "" || payload.SubscriptionID != subscriptionID {
		return Delivery{}, errUnsafeDeliveryPayload
	}
	plaintext, err := s.protector.Decrypt(persistence.PushTokenCiphertext{KeyID: row.EncryptionKeyID, Ciphertext: row.TokenCiphertext})
	if err != nil || len(plaintext) == 0 {
		return Delivery{}, errTokenDecryption
	}
	quiet, err := quietHoursFromRow(row)
	if err != nil {
		return Delivery{}, err
	}
	return Delivery{ID: id, TokenID: tokenID, SubscriptionID: subscriptionID, Type: row.NotificationType, EntityID: payload.EntityID, DeepLink: payload.DeepLink, PushToken: string(plaintext), ExpiresAt: row.ExpiresAt.Time.UTC(), Attempts: int(row.Attempts), QuietHours: quiet}, nil
}

func (s *PostgresStore) MarkSent(ctx context.Context, id, ticketID string, at time.Time) error {
	if strings.TrimSpace(ticketID) == "" || len(ticketID) > 256 || strings.ContainsAny(ticketID, "\r\n\x00") {
		return errors.New("invalid provider ticket ID")
	}
	parsedID, err := parseUUID(id)
	if err != nil {
		return err
	}
	changed, err := s.queries.MarkNotificationDeliverySent(ctx, sqlcgen.MarkNotificationDeliverySentParams{ID: parsedID, ProviderTicketID: pgText(ticketID), SentAt: pgTimestamp(at)})
	return stateChange(changed, err)
}

func (s *PostgresStore) MarkRetry(ctx context.Context, id string, next time.Time, code string, at time.Time) error {
	if !safeErrorCode(code) || !next.After(time.Time{}) {
		return errors.New("invalid notification retry transition")
	}
	parsedID, err := parseUUID(id)
	if err != nil {
		return err
	}
	changed, err := s.queries.MarkNotificationDeliveryRetry(ctx, sqlcgen.MarkNotificationDeliveryRetryParams{ID: parsedID, NextAttemptAt: pgTimestamp(next), ErrorCode: pgText(code), UpdatedAt: pgTimestamp(at)})
	return stateChange(changed, err)
}

func (s *PostgresStore) MarkExpired(ctx context.Context, id string, at time.Time) error {
	parsedID, err := parseUUID(id)
	if err != nil {
		return err
	}
	changed, err := s.queries.MarkNotificationDeliveryExpired(ctx, sqlcgen.MarkNotificationDeliveryExpiredParams{ID: parsedID, UpdatedAt: pgTimestamp(at)})
	return stateChange(changed, err)
}

func (s *PostgresStore) MarkFailed(ctx context.Context, id, code string, at time.Time) error {
	if !safeErrorCode(code) {
		return errors.New("invalid notification error code")
	}
	parsedID, err := parseUUID(id)
	if err != nil {
		return err
	}
	changed, err := s.queries.MarkNotificationDeliveryFailed(ctx, sqlcgen.MarkNotificationDeliveryFailedParams{ID: parsedID, ErrorCode: pgText(code), UpdatedAt: pgTimestamp(at)})
	return stateChange(changed, err)
}

func (s *PostgresStore) DisableToken(ctx context.Context, id, code string, at time.Time) error {
	if code != "invalid_token" {
		return errors.New("invalid push token disable code")
	}
	parsedID, err := parseUUID(id)
	if err != nil {
		return err
	}
	return s.queries.DisablePushToken(ctx, sqlcgen.DisablePushTokenParams{ID: parsedID, DisabledReason: pgText(code), DisabledAt: pgTimestamp(at)})
}

// RecordReceipt persists a sanitized provider result and atomically disables a
// token only for an invalid-token receipt tied to its own provider ticket.
func (s *PostgresStore) RecordReceipt(ctx context.Context, receipt Receipt, at time.Time) error {
	if s == nil || s.beginReceipt == nil || !validReceipt(receipt) {
		return ErrInvalidNotificationReceipt
	}
	tx, err := s.beginReceipt(ctx)
	if err != nil {
		return fmt.Errorf("begin notification receipt transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	if _, err := tx.RecordNotificationReceipt(ctx, sqlcgen.RecordNotificationReceiptParams{ProviderTicketID: receipt.TicketID, ReceivedAt: pgTimestamp(receipt.ReceivedAt), Status: receipt.Status, ErrorCode: pgTextOptional(receipt.ErrorCode), SafeDetail: pgTextOptional(receipt.SafeDetail), ProcessedAt: pgTimestamp(at)}); err != nil {
		return fmt.Errorf("record notification receipt: %w", err)
	}
	if receipt.Status == "error" && receipt.ErrorCode == "invalid_token" {
		if _, err := tx.DisableTokenForInvalidReceipt(ctx, sqlcgen.DisableTokenForInvalidReceiptParams{ProviderTicketID: pgText(receipt.TicketID), DisabledAt: pgTimestamp(at)}); err != nil {
			return fmt.Errorf("disable invalid receipt token: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit notification receipt transaction: %w", err)
	}
	committed = true
	return nil
}

func stateChange(changed int64, err error) error {
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrNotificationStateConflict
	}
	return nil
}

func safeErrorCode(code string) bool {
	switch code {
	case "invalid_token", "push_unavailable", "quiet_hours", "invalid_quiet_hours", "token_decryption_failed", "invalid_delivery_payload":
		return true
	default:
		return false
	}
}

func validReceipt(receipt Receipt) bool {
	if strings.TrimSpace(receipt.TicketID) == "" || len(receipt.TicketID) > 256 || strings.ContainsAny(receipt.TicketID, "\r\n\x00") || (receipt.Status != "ok" && receipt.Status != "error") || receipt.ReceivedAt.IsZero() || len(receipt.SafeDetail) > 240 || strings.ContainsAny(receipt.SafeDetail, "\r\n\x00") {
		return false
	}
	if receipt.Status == "ok" {
		return receipt.ErrorCode == "" && receipt.SafeDetail == ""
	}
	return receipt.ErrorCode == "invalid_token" || receipt.ErrorCode == "provider_error"
}

func quietHoursFromRow(row sqlcgen.ClaimNotificationDeliveriesRow) (*QuietHours, error) {
	if !row.QuietTimeZone.Valid && !row.QuietStart.Valid && !row.QuietEnd.Valid {
		return nil, nil
	}
	if !row.QuietTimeZone.Valid || !row.QuietStart.Valid || !row.QuietEnd.Valid {
		return nil, errors.New("incomplete quiet hours")
	}
	format := func(value pgtype.Time) string {
		minutes := value.Microseconds / int64(time.Minute/time.Microsecond)
		return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
	}
	return &QuietHours{Start: format(row.QuietStart), End: format(row.QuietEnd), TimeZone: row.QuietTimeZone.String}, nil
}

func pgTimestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}
func pgText(value string) pgtype.Text         { return pgtype.Text{String: value, Valid: true} }
func pgTextOptional(value string) pgtype.Text { return pgtype.Text{String: value, Valid: value != ""} }

func uuidString(value pgtype.UUID) (string, error) {
	if !value.Valid {
		return "", errors.New("invalid notification UUID")
	}
	return value.String(), nil
}

func parseUUID(value string) (pgtype.UUID, error) {
	var parsed pgtype.UUID
	if err := parsed.Scan(value); err != nil || !parsed.Valid {
		return pgtype.UUID{}, errors.New("invalid notification UUID")
	}
	return parsed, nil
}
