package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	notificationworker "github.com/toozej/tabi-transit/services/notification-worker"
)

func TestRunFailsClosedOnlyWhenDeliveryIsEnabled(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	if code := run(notificationworker.RuntimeConfig{}, nil, &stderr); code != 0 {
		t.Fatalf("disabled exit code = %d", code)
	}
	if code := run(notificationworker.RuntimeConfig{Enabled: true}, nil, &stderr); code != 2 || !strings.Contains(stderr.String(), "cannot start") {
		t.Fatalf("enabled exit code = %d, stderr = %q", code, stderr.String())
	}
	if code := run(notificationworker.RuntimeConfig{}, errors.New("bad configuration"), &stderr); code != 2 || !strings.Contains(stderr.String(), "bad configuration") {
		t.Fatalf("configuration error = %d, stderr = %q", code, stderr.String())
	}
}
