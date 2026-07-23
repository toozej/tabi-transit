package main

import (
	"context"
	"fmt"
	"os"

	"github.com/toozej/tabi-transit/services/realtime-poller"
)

// Wiring the production persistence store is intentionally deferred until its
// transactional realtime implementation is supplied by the persistence owner.
func main() {
	if _, err := poller.DefaultConfig(); err != nil {
		if err == poller.ErrDisabled {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	_ = context.Background()
	fmt.Fprintln(os.Stderr, "realtime-poller persistence wiring is not configured")
	os.Exit(2)
}
