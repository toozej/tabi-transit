// tabi-healthcheck is a dependency-free HTTP readiness probe for Compose.
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: tabi-healthcheck http://host/path")
		os.Exit(2)
	}
	target, err := url.ParseRequestURI(os.Args[1])
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		fmt.Fprintln(os.Stderr, "healthcheck target must be an absolute http(s) URL")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck request:", err)
		os.Exit(2)
	}
	response, err := (&http.Client{Timeout: 4 * time.Second}).Do(request)
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck request failed:", err)
		os.Exit(1)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Fprintln(os.Stderr, "healthcheck returned:", response.Status)
		os.Exit(1)
	}
}
