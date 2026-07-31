// tabi-healthcheck is a dependency-free HTTP readiness probe for Compose.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:], &http.Client{Timeout: 4 * time.Second}, os.Stderr))
}

func run(args []string, client *http.Client, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: tabi-healthcheck http://host/path")
		return 2
	}
	target, err := url.ParseRequestURI(args[0])
	if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
		fmt.Fprintln(stderr, "healthcheck target must be an absolute http(s) URL")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		fmt.Fprintln(stderr, "healthcheck request:", err)
		return 2
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Fprintln(stderr, "healthcheck request failed:", err)
		return 1
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Fprintln(stderr, "healthcheck returned:", response.Status)
		return 1
	}
	return 0
}
