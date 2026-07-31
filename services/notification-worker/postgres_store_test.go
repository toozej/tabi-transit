package notificationworker

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/toozej/tabi-transit/internal/persistence"
	"github.com/toozej/tabi-transit/internal/persistence/sqlcgen"
)

type fakeNotificationQueries struct {
	rows       []sqlcgen.ClaimNotificationDeliveriesRow
	failedCode string
	retry      sqlcgen.MarkNotificationDeliveryRetryParams
}

func (q *fakeNotificationQueries) ClaimNotificationDeliveries(_ context.Context, _ sqlcgen.ClaimNotificationDeliveriesParams) ([]sqlcgen.ClaimNotificationDeliveriesRow, error) {
	return q.rows, nil
}
func (*fakeNotificationQueries) DisablePushToken(context.Context, sqlcgen.DisablePushTokenParams) error {
	return nil
}
func (*fakeNotificationQueries) ExpirePendingDeliveries(context.Context, pgtype.Timestamptz) (int64, error) {
	return 0, nil
}
func (*fakeNotificationQueries) MarkNotificationDeliverySent(context.Context, sqlcgen.MarkNotificationDeliverySentParams) (int64, error) {
	return 1, nil
}
func (q *fakeNotificationQueries) MarkNotificationDeliveryRetry(_ context.Context, value sqlcgen.MarkNotificationDeliveryRetryParams) (int64, error) {
	q.retry = value
	return 1, nil
}
func (*fakeNotificationQueries) MarkNotificationDeliveryExpired(context.Context, sqlcgen.MarkNotificationDeliveryExpiredParams) (int64, error) {
	return 1, nil
}
func (q *fakeNotificationQueries) MarkNotificationDeliveryFailed(_ context.Context, value sqlcgen.MarkNotificationDeliveryFailedParams) (int64, error) {
	q.failedCode = value.ErrorCode.String
	return 1, nil
}

type fakeReceiptTx struct {
	recorded  sqlcgen.RecordNotificationReceiptParams
	disabled  bool
	committed bool
	rolled    bool
	err       error
}

func (t *fakeReceiptTx) RecordNotificationReceipt(_ context.Context, value sqlcgen.RecordNotificationReceiptParams) (string, error) {
	t.recorded = value
	return value.ProviderTicketID, t.err
}
func (t *fakeReceiptTx) DisableTokenForInvalidReceipt(context.Context, sqlcgen.DisableTokenForInvalidReceiptParams) (int64, error) {
	t.disabled = true
	return 1, nil
}
func (t *fakeReceiptTx) Commit(context.Context) error   { t.committed = true; return nil }
func (t *fakeReceiptTx) Rollback(context.Context) error { t.rolled = true; return nil }

func testUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var result pgtype.UUID
	if err := result.Scan(value); err != nil {
		t.Fatal(err)
	}
	return result
}

func testStore(t *testing.T, queries *fakeNotificationQueries, tx *fakeReceiptTx) *PostgresStore {
	t.Helper()
	key := make([]byte, 32)
	protector, err := persistence.NewAESGCMTokenProtector("test-key", key)
	if err != nil {
		t.Fatal(err)
	}
	return &PostgresStore{queries: queries, protector: protector, claimLease: time.Minute, beginReceipt: func(context.Context) (receiptTransaction, error) { return tx, nil }}
}

func TestPostgresStoreClaimDecryptsAfterClaim(t *testing.T) {
	t.Parallel()
	queries := &fakeNotificationQueries{}
	store := testStore(t, queries, &fakeReceiptTx{})
	ciphertext, err := store.protector.Encrypt([]byte("opaque-token"))
	if err != nil {
		t.Fatal(err)
	}
	subscriptionID := "123e4567-e89b-12d3-a456-426614174002"
	queries.rows = []sqlcgen.ClaimNotificationDeliveriesRow{{
		ID: testUUID(t, "123e4567-e89b-12d3-a456-426614174000"), PushTokenID: testUUID(t, "123e4567-e89b-12d3-a456-426614174001"), SubscriptionID: testUUID(t, subscriptionID),
		NotificationType: "service_alert", Payload: []byte(`{"entityId":"alert-1","deepLink":"tabi://alerts/alert-1","subscriptionId":"123e4567-e89b-12d3-a456-426614174002"}`),
		ExpiresAt: pgTimestamp(time.Now().Add(time.Hour)), Attempts: 1, ClaimToken: testUUID(t, "123e4567-e89b-12d3-a456-426614174003"), DedupeKey: "dedupe-key", TokenCiphertext: ciphertext.Ciphertext, EncryptionKeyID: ciphertext.KeyID,
		QuietStart: pgtype.Time{Microseconds: 22 * int64(time.Hour/time.Microsecond), Valid: true}, QuietEnd: pgtype.Time{Microseconds: 7 * int64(time.Hour/time.Microsecond), Valid: true}, QuietTimeZone: pgText("America/Los_Angeles"),
	}}
	got, err := store.Claim(context.Background(), time.Now(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].PushToken != "opaque-token" || got[0].ClaimToken == "" || got[0].IdempotencyKey != "dedupe-key" || got[0].QuietHours == nil || got[0].QuietHours.Start != "22:00" {
		t.Fatalf("claim = %#v", got)
	}
}

func TestPostgresStoreClaimMakesUndecryptableRowsTerminal(t *testing.T) {
	t.Parallel()
	subscriptionID := "123e4567-e89b-12d3-a456-426614174002"
	queries := &fakeNotificationQueries{rows: []sqlcgen.ClaimNotificationDeliveriesRow{{
		ID: testUUID(t, "123e4567-e89b-12d3-a456-426614174000"), PushTokenID: testUUID(t, "123e4567-e89b-12d3-a456-426614174001"), SubscriptionID: testUUID(t, subscriptionID),
		Payload: []byte(`{"entityId":"alert-1","deepLink":"tabi://alerts/alert-1","subscriptionId":"123e4567-e89b-12d3-a456-426614174002"}`), ExpiresAt: pgTimestamp(time.Now().Add(time.Hour)), ClaimToken: testUUID(t, "123e4567-e89b-12d3-a456-426614174003"), TokenCiphertext: []byte("not-a-ciphertext"), EncryptionKeyID: "test-key",
	}}}
	store := testStore(t, queries, &fakeReceiptTx{})
	got, err := store.Claim(context.Background(), time.Now(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 || queries.failedCode != "token_decryption_failed" {
		t.Fatalf("deliveries=%#v failed=%q", got, queries.failedCode)
	}
}

func TestPostgresStoreReceiptInvalidatesOnlyInTransaction(t *testing.T) {
	t.Parallel()
	tx := &fakeReceiptTx{}
	store := testStore(t, &fakeNotificationQueries{}, tx)
	err := store.RecordReceipt(context.Background(), Receipt{TicketID: "ticket-safe", Status: "error", ErrorCode: "invalid_token", SafeDetail: "unregistered", ReceivedAt: time.Now()}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !tx.disabled || !tx.committed || tx.rolled || tx.recorded.SafeDetail.String != "unregistered" {
		t.Fatalf("transaction=%#v", tx)
	}
}

func TestPostgresStoreReceiptRollsBackOnWriteFailure(t *testing.T) {
	t.Parallel()
	tx := &fakeReceiptTx{err: errors.New("db failure")}
	store := testStore(t, &fakeNotificationQueries{}, tx)
	err := store.RecordReceipt(context.Background(), Receipt{TicketID: "ticket-safe", Status: "ok", ReceivedAt: time.Now()}, time.Now())
	if err == nil || !tx.rolled || tx.committed {
		t.Fatalf("err=%v transaction=%#v", err, tx)
	}
}

func TestPostgresStoreRejectsUnsafeTransitionValues(t *testing.T) {
	t.Parallel()
	store := testStore(t, &fakeNotificationQueries{}, &fakeReceiptTx{})
	if err := store.MarkRetry(context.Background(), "not-a-uuid", "123e4567-e89b-12d3-a456-426614174003", time.Now(), "provider body: opaque-token", time.Now()); err == nil {
		t.Fatal("unsafe error code accepted")
	}
	if err := store.RecordReceipt(context.Background(), Receipt{TicketID: "ticket\nsecret", Status: "ok", ReceivedAt: time.Now()}, time.Now()); !errors.Is(err, ErrInvalidNotificationReceipt) {
		t.Fatalf("receipt error = %v", err)
	}
}

func TestLoadRuntimeConfigFailsClosed(t *testing.T) {
	t.Setenv("TABI_NOTIFICATION_DELIVERY_ENABLED", "true")
	t.Setenv("TABI_NOTIFICATION_TOKEN_KEY_ID", "")
	if _, err := LoadRuntimeConfig(); err == nil {
		t.Fatal("enabled delivery without key ID was accepted")
	}
	t.Setenv("TABI_NOTIFICATION_TOKEN_KEY_ID", "test-key")
	t.Setenv("TABI_NOTIFICATION_TOKEN_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 31)))
	if _, err := LoadRuntimeConfig(); err == nil {
		t.Fatal("short key was accepted")
	}
	t.Setenv("TABI_NOTIFICATION_TOKEN_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	configured, err := LoadRuntimeConfig()
	if err != nil || !configured.Enabled || configured.Protector == nil {
		t.Fatalf("config=%#v err=%v", configured, err)
	}
}
