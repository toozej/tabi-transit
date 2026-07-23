package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretFileTakesPrecedence(t *testing.T) {
	p := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(p, []byte(" value\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TAB_TEST_SECRET", "fallback")
	t.Setenv("TAB_TEST_SECRET_FILE", p)
	v, err := Secret("TAB_TEST_SECRET")
	if err != nil || v != "value" {
		t.Fatalf("got %q, %v", v, err)
	}
}
func TestLoadRejectsInvalidBound(t *testing.T) {
	t.Setenv("TABI_RATE_LIMIT_REQUESTS", "0")
	if _, err := Load(); err == nil {
		t.Fatal("accepted zero")
	}
}
