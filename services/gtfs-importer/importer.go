// Package importer orchestrates a fixture-first, atomic static GTFS import.
package importer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/toozej/tabi-transit/internal/sources/gtfs"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var ErrDisabled = errors.New("GTFS source endpoint is not configured")

type Config struct {
	SourceID, Endpoint, EndpointFile, ArchiveSHA256 string
	AllowedHosts                                    []string
	Timeout                                         time.Duration
	Policy                                          gtfs.ArchivePolicy
}

func DefaultPolicy() gtfs.ArchivePolicy { return gtfs.DefaultArchivePolicy() }
func (c Config) Validate() error {
	if c.SourceID == "" {
		return errors.New("GTFS_SOURCE_ID is required")
	}
	if c.Endpoint == "" && c.EndpointFile != "" {
		b, e := os.ReadFile(c.EndpointFile)
		if e != nil {
			return fmt.Errorf("GTFS_ENDPOINT_FILE: %w", e)
		}
		c.Endpoint = strings.TrimSpace(string(b))
	}
	if c.Endpoint == "" {
		return ErrDisabled
	}
	u, e := url.Parse(c.Endpoint)
	if e != nil || u.Scheme != "https" || u.Host == "" {
		return errors.New("GTFS_ENDPOINT must be an HTTPS URL")
	}
	ok := false
	for _, h := range c.AllowedHosts {
		if strings.EqualFold(u.Hostname(), h) {
			ok = true
		}
	}
	if !ok {
		return errors.New("GTFS_ENDPOINT host is not allowlisted")
	}
	if c.Timeout <= 0 {
		return errors.New("GTFS_HTTP_TIMEOUT must be positive")
	}
	return nil
}
func (c Config) resolved() (Config, error) {
	if c.Endpoint == "" && c.EndpointFile != "" {
		b, err := os.ReadFile(c.EndpointFile)
		if err != nil {
			return c, fmt.Errorf("GTFS_ENDPOINT_FILE: %w", err)
		}
		c.Endpoint = strings.TrimSpace(string(b))
	}
	return c, c.Validate()
}

type Store interface {
	Import(ctx context.Context, sourceID, label, digest string, fetchedAt time.Time, feed gtfs.Feed) (alreadyActive bool, err error)
	RecordFailure(ctx context.Context, sourceID, code string, at time.Time) error
}
type Reporter interface {
	ImportResult(sourceID, outcome, code string, count int)
}
type Service struct {
	Config     Config
	HTTPClient *http.Client
	Store      Store
	Reporter   Reporter
	Clock      func() time.Time
}

func (s Service) Run(ctx context.Context) error {
	config, err := s.Config.resolved()
	if err != nil {
		return err
	}
	now := time.Now
	if s.Clock != nil {
		now = s.Clock
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, config.Endpoint, nil)
	if e != nil {
		return e
	}
	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: config.Timeout}
	}
	resp, e := client.Do(req)
	if e != nil {
		return s.fail(ctx, "fetch_failed", now, e)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return s.fail(ctx, "fetch_status", now, fmt.Errorf("unexpected status %d", resp.StatusCode))
	}
	feed, e := gtfs.Read(resp.Body, config.Policy)
	if e != nil {
		return s.fail(ctx, "validation_failed", now, e)
	}
	if config.ArchiveSHA256 != "" && config.ArchiveSHA256 != feed.SHA256 {
		return s.fail(ctx, "checksum_mismatch", now, errors.New("archive checksum mismatch"))
	}
	if s.Store == nil {
		return s.fail(ctx, "load_failed", now, errors.New("GTFS store is not configured"))
	}
	active, e := s.Store.Import(ctx, s.Config.SourceID, feed.SHA256[:16], feed.SHA256, now(), feed)
	if e != nil {
		return s.fail(ctx, "load_failed", now, e)
	}
	if s.Reporter != nil {
		s.Reporter.ImportResult(s.Config.SourceID, "succeeded", "", len(feed.Stops))
	}
	_ = active
	return nil
}
func (s Service) fail(ctx context.Context, code string, at func() time.Time, err error) error {
	if s.Store != nil {
		_ = s.Store.RecordFailure(ctx, s.Config.SourceID, code, at())
	}
	if s.Reporter != nil {
		s.Reporter.ImportResult(s.Config.SourceID, "failed", code, 0)
	}
	return fmt.Errorf("gtfs import %s: %w", code, err)
}
func SHA256(r io.Reader) (string, error) {
	h := sha256.New()
	_, e := io.Copy(h, r)
	return hex.EncodeToString(h.Sum(nil)), e
}
