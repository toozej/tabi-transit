package main

import (
	"fmt"
	"io"
	"os"

	notificationworker "github.com/toozej/tabi-transit/services/notification-worker"
)

func main() {
	runtime, err := notificationworker.LoadRuntimeConfig()
	if code := run(runtime, err, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

func run(runtime notificationworker.RuntimeConfig, err error, stderr io.Writer) int {
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if !runtime.Enabled {
		return 0
	}
	// D-017 and D-018 deliberately prevent construction of a provider client,
	// database store, or network loop. A true flag without that approved
	// composition fails closed instead of silently sending anything.
	fmt.Fprintln(stderr, "notification delivery cannot start until approved provider composition is added")
	return 2
}
