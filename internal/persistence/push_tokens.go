package persistence

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

var ErrInvalidPushTokenKey = errors.New("push token encryption key must be 32 bytes")
var ErrInvalidPushTokenCiphertext = errors.New("invalid push token ciphertext")

// PushTokenCiphertext is the only token representation safe to persist. The
// nonce is prefixed to Ciphertext so rotation can remain an application-level
// operation keyed by KeyID. It intentionally has no String method.
type PushTokenCiphertext struct {
	KeyID      string
	Ciphertext []byte
}

// PushTokenProtector isolates encryption at rest from registration and worker
// code. Production key loading remains at the binary boundary via *_FILE.
type PushTokenProtector interface {
	Encrypt([]byte) (PushTokenCiphertext, error)
	Decrypt(PushTokenCiphertext) ([]byte, error)
}

type AESGCMTokenProtector struct {
	keyID string
	aead  cipher.AEAD
	rand  io.Reader
}

func NewAESGCMTokenProtector(keyID string, key []byte) (*AESGCMTokenProtector, error) {
	if len(key) != 32 {
		return nil, ErrInvalidPushTokenKey
	}
	if keyID == "" || len(keyID) > 128 {
		return nil, errors.New("invalid push token encryption key ID")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create push token cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create push token AEAD: %w", err)
	}
	return &AESGCMTokenProtector{keyID: keyID, aead: aead, rand: rand.Reader}, nil
}

func (p *AESGCMTokenProtector) Encrypt(token []byte) (PushTokenCiphertext, error) {
	if p == nil || p.aead == nil || len(token) == 0 {
		return PushTokenCiphertext{}, errors.New("cannot encrypt empty push token")
	}
	nonce := make([]byte, p.aead.NonceSize())
	if _, err := io.ReadFull(p.rand, nonce); err != nil {
		return PushTokenCiphertext{}, fmt.Errorf("generate push token nonce: %w", err)
	}
	return PushTokenCiphertext{KeyID: p.keyID, Ciphertext: p.aead.Seal(nonce, nonce, token, []byte(p.keyID))}, nil
}

func (p *AESGCMTokenProtector) Decrypt(value PushTokenCiphertext) ([]byte, error) {
	if p == nil || p.aead == nil || value.KeyID != p.keyID || len(value.Ciphertext) <= p.aead.NonceSize() {
		return nil, ErrInvalidPushTokenCiphertext
	}
	nonce, ciphertext := value.Ciphertext[:p.aead.NonceSize()], value.Ciphertext[p.aead.NonceSize():]
	plain, err := p.aead.Open(nil, nonce, ciphertext, []byte(value.KeyID))
	if err != nil {
		return nil, ErrInvalidPushTokenCiphertext
	}
	return plain, nil
}
