package main

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRunValidatesTargetAndResponse(t *testing.T) {
	t.Parallel()
	client := &http.Client{Transport: roundTrip(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", request.Method)
		}
		return response(http.StatusNoContent), nil
	})}

	var stderr bytes.Buffer
	if code := run([]string{"https://healthcheck.example/ready"}, client, &stderr); code != 0 {
		t.Fatalf("success exit code = %d, stderr = %q", code, stderr.String())
	}
	if code := run([]string{"relative/path"}, client, &stderr); code != 2 || !strings.Contains(stderr.String(), "absolute") {
		t.Fatalf("invalid target = %d, %q", code, stderr.String())
	}
}

func TestRunFailsForUnhealthyResponse(t *testing.T) {
	t.Parallel()
	client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return response(http.StatusServiceUnavailable), nil
	})}

	var stderr bytes.Buffer
	if code := run([]string{"https://healthcheck.example/ready"}, client, &stderr); code != 1 || !strings.Contains(stderr.String(), "Service Unavailable") {
		t.Fatalf("unhealthy response = %d, %q", code, stderr.String())
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func response(status int) *http.Response {
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}
}
