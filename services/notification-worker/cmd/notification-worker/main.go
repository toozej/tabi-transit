package main

import (
	"fmt"
	"os"

	notificationworker "github.com/toozej/tabi-transit/services/notification-worker"
)

func main() {
	runtime, err := notificationworker.LoadRuntimeConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if !runtime.Enabled {
		return
	}
	// D-017 and D-018 deliberately prevent construction of a provider client,
	// database store, or network loop. A true flag without that approved
	// composition fails closed instead of silently sending anything.
	fmt.Fprintln(os.Stderr, "notification delivery cannot start until approved provider composition is added")
	os.Exit(2)
}
