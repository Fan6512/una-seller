package notifications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// Only fixed phase names, sanitized error classes and durations are emitted.
// No payload, identifiers, token, raw SQL/driver error, or source IP is logged.
// This trace belongs to one authenticated live admission, not to workers.
type ackTrace struct {
	started, changed time.Time
	stage            string
	durations        []string
}

func newACKTrace() *ackTrace {
	now := time.Now()
	return &ackTrace{started: now, changed: now, stage: "acquire"}
}

func (t *ackTrace) next(stage string) {
	if t == nil {
		return
	}
	now := time.Now()
	t.durations = append(t.durations, fmt.Sprintf("%s_ms=%d", t.stage, now.Sub(t.changed).Milliseconds()))
	t.stage, t.changed = stage, now
}

func ackFailureClass(err error) string {
	var postgres *pgconn.PgError
	switch {
	case err == nil:
		return "committed"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case pgconn.Timeout(err):
		return "driver_timeout"
	case errors.As(err, &postgres) && postgres.Code == "57014":
		// query_canceled may be a statement timeout or an explicit cancel;
		// it does not by itself prove the request deadline was exceeded.
		return "query_cancelled"
	case errors.Is(err, ErrScope):
		return "scope_rejected"
	case errors.Is(err, ErrScopeUnavailable):
		return "scope_unavailable"
	case errors.Is(err, ErrCapacity):
		return "capacity"
	default:
		return "storage_error"
	}
}

func (t *ackTrace) report(err error) {
	elapsed := time.Since(t.started)
	if err == nil && elapsed < 200*time.Millisecond {
		return
	}
	// The final phase may include transaction cleanup; timings are client-side
	// wall time, not individual server execution times or proof of commit.
	last := t.stage
	t.next("finished")
	log.Printf("notification admission outcome=%s final_stage=%s elapsed_ms=%d %s",
		ackFailureClass(err), last, elapsed.Milliseconds(), strings.Join(t.durations, " "))
}
