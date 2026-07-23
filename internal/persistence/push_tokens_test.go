package persistence

import (
	"bytes"
	"errors"
	"testing"
)

func TestAESGCMTokenProtectorRoundTripAndTamperResistance(t *testing.T) {
	t.Parallel()
	key := bytes.Repeat([]byte{7}, 32)
	protector, err := NewAESGCMTokenProtector("test-key-1", key)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := protector.Encrypt([]byte("ExponentPushToken[redacted]"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(sealed.Ciphertext, []byte("ExponentPushToken")) {
		t.Fatal("ciphertext contains plaintext token")
	}
	plain, err := protector.Decrypt(sealed)
	if err != nil || string(plain) != "ExponentPushToken[redacted]" {
		t.Fatalf("round trip = %q, %v", plain, err)
	}
	sealed.Ciphertext[len(sealed.Ciphertext)-1] ^= 1
	if _, err := protector.Decrypt(sealed); !errors.Is(err, ErrInvalidPushTokenCiphertext) {
		t.Fatalf("tampered ciphertext error = %v", err)
	}
}

func TestAESGCMTokenProtectorRejectsInvalidKeys(t *testing.T) {
	t.Parallel()
	if _, err := NewAESGCMTokenProtector("key", []byte("short")); !errors.Is(err, ErrInvalidPushTokenKey) {
		t.Fatalf("invalid key error = %v", err)
	}
}
