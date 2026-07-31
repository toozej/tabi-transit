package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretPrefersTrimmedFileValue(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "database-url")
	if err := os.WriteFile(path, []byte("  postgres://from-file  \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_DATABASE_URL", "postgres://environment")
	t.Setenv("TEST_DATABASE_URL_FILE", path)

	got, err := secret("TEST_DATABASE_URL")
	if err != nil || got != "postgres://from-file" {
		t.Fatalf("secret = %q, %v", got, err)
	}
}

func TestSecretRejectsEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(path, []byte(" \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_EMPTY_FILE", path)

	if _, err := secret("TEST_EMPTY"); err == nil {
		t.Fatal("empty file accepted")
	}
}
