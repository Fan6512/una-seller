package notifications

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestACKFailureClass(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want string
	}{
		{nil, "committed"},
		{fmt.Errorf("wrapped: %w", context.DeadlineExceeded), "deadline"},
		{context.Canceled, "cancelled"},
		{fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "57014"}), "query_cancelled"},
		{&pgconn.PgError{Code: "23505"}, "storage_error"},
		{ErrScope, "scope_rejected"},
		{ErrScopeUnavailable, "scope_unavailable"},
		{ErrCapacity, "capacity"},
		{errors.New("private driver message"), "storage_error"},
	} {
		if got := ackFailureClass(tc.err); got != tc.want {
			t.Fatalf("classification=%q want=%q", got, tc.want)
		}
	}
}
