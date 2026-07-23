package notificationworker

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/toozej/tabi-transit/internal/config"
	"github.com/toozej/tabi-transit/internal/persistence"
)

var ErrNotificationDeliveryDisabled = errors.New("notification delivery is disabled")

// RuntimeConfig is intentionally insufficient to send a push. It makes the
// feature gate and encryption prerequisite explicit while D-017/D-018 prevent
// composition of a real gateway.
type RuntimeConfig struct {
	Enabled   bool
	Protector persistence.PushTokenProtector
}

// LoadRuntimeConfig never loads key material when the clearly named feature
// flag is absent or false. When enabled, a key ID and a 32-byte base64 key must
// be supplied through the normal *_FILE secret convention.
func LoadRuntimeConfig() (RuntimeConfig, error) {
	rawEnabled := strings.TrimSpace(os.Getenv("TABI_NOTIFICATION_DELIVERY_ENABLED"))
	if rawEnabled == "" || rawEnabled == "false" {
		return RuntimeConfig{}, nil
	}
	if rawEnabled != "true" {
		return RuntimeConfig{}, errors.New("TABI_NOTIFICATION_DELIVERY_ENABLED must be true or false")
	}
	keyID := strings.TrimSpace(os.Getenv("TABI_NOTIFICATION_TOKEN_KEY_ID"))
	if keyID == "" {
		return RuntimeConfig{}, errors.New("TABI_NOTIFICATION_TOKEN_KEY_ID is required when notification delivery is enabled")
	}
	rawKey, err := config.Secret("TABI_NOTIFICATION_TOKEN_ENCRYPTION_KEY")
	if err != nil || rawKey == "" {
		return RuntimeConfig{}, errors.New("TABI_NOTIFICATION_TOKEN_ENCRYPTION_KEY or TABI_NOTIFICATION_TOKEN_ENCRYPTION_KEY_FILE is required when notification delivery is enabled")
	}
	key, err := base64.StdEncoding.DecodeString(rawKey)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("decode notification token encryption key: %w", err)
	}
	protector, err := persistence.NewAESGCMTokenProtector(keyID, key)
	if err != nil {
		return RuntimeConfig{}, err
	}
	return RuntimeConfig{Enabled: true, Protector: protector}, nil
}
