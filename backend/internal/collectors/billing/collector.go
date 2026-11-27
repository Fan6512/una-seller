// Package billing implements the built-in collector.meli.billing lifecycle.
// Finance owns authorization, batching, persisted evidence and net arithmetic.
package billing

import (
	"context"
	"log"
	"time"
)

const ID = "collector.meli.billing"
const Interval = 2 * time.Hour
const MaxBatch = 10
const MaxAttempts = 3

// Run performs one bounded reconciliation pass per interval, never on the
// order sync, notification ACK or interactive profit read path.
func Run(ctx context.Context, logger *log.Logger, refresh func(context.Context, time.Time) error) {
	ticker := time.NewTicker(Interval)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		pass, cancel := context.WithTimeout(ctx, 2*time.Minute)
		err := refresh(pass, time.Now().UTC())
		cancel()
		if err != nil && ctx.Err() == nil {
			// Never log provider bodies, tokens or order identifiers.
			logger.Printf("%s pass unavailable; retry at next interval", ID)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
