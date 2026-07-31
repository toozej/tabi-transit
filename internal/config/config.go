// Package config loads process configuration without exposing secret values.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddress string
	// DatabaseURL is only consumed by composition roots and is never logged.
	// It supports TABI_DATABASE_URL_FILE through Secret.
	DatabaseURL string
	API         PublicAPI
	RateLimit   RateLimit
}

type PublicAPI struct {
	Version               string
	MinimumAppVersion     string
	StaleThresholdSeconds int
	StaticFeedVersion     string
	StaticFeedPublishedAt time.Time
}

type RateLimit struct {
	Requests int
	Window   time.Duration
}

func Load() (Config, error) {
	listenAddress := value("TABI_API_LISTEN_ADDRESS", "")
	if listenAddress == "" {
		listenAddress = value("HTTP_LISTEN_ADDR", ":8080")
	}
	c := Config{ListenAddress: listenAddress, API: PublicAPI{
		Version: value("TABI_API_VERSION", "0.1.0"), MinimumAppVersion: value("TABI_MINIMUM_APP_VERSION", "0.1.0"),
		StaleThresholdSeconds: 90, StaticFeedVersion: value("TABI_STATIC_FEED_VERSION", "unknown"),
		StaticFeedPublishedAt: time.Unix(0, 0).UTC(),
	}, RateLimit: RateLimit{Requests: 120, Window: time.Minute}}
	var err error
	if c.API.StaleThresholdSeconds, err = positiveInt("TABI_STALE_THRESHOLD_SECONDS", c.API.StaleThresholdSeconds); err != nil {
		return Config{}, err
	}
	if c.RateLimit.Requests, err = positiveInt("TABI_RATE_LIMIT_REQUESTS", c.RateLimit.Requests); err != nil {
		return Config{}, err
	}
	if seconds, e := positiveInt("TABI_RATE_LIMIT_WINDOW_SECONDS", 60); e != nil {
		return Config{}, e
	} else {
		// Check before conversion: a large valid int can wrap Duration.
		if int64(seconds) > int64(time.Duration(1<<63-1)/time.Second) {
			return Config{}, fmt.Errorf("TABI_RATE_LIMIT_WINDOW_SECONDS is too large")
		}
		c.RateLimit.Window = time.Duration(seconds) * time.Second
	}
	if raw := strings.TrimSpace(os.Getenv("TABI_STATIC_FEED_PUBLISHED_AT")); raw != "" {
		if c.API.StaticFeedPublishedAt, err = time.Parse(time.RFC3339, raw); err != nil {
			return Config{}, fmt.Errorf("TABI_STATIC_FEED_PUBLISHED_AT: %w", err)
		}
	}
	if c.DatabaseURL, err = Secret("TABI_DATABASE_URL"); err != nil {
		return Config{}, err
	}
	return c, nil
}

// Secret reads support the conventional NAME_FILE override. Only callers that
// explicitly request a secret receive its value.
func Secret(name string) (string, error) {
	if path := strings.TrimSpace(os.Getenv(name + "_FILE")); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("%s_FILE: %w", name, err)
		}
		v := strings.TrimSpace(string(b))
		if v == "" {
			return "", errors.New(name + "_FILE is empty")
		}
		return v, nil
	}
	return strings.TrimSpace(os.Getenv(name)), nil
}
func value(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}
func positiveInt(name string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return n, nil
}
