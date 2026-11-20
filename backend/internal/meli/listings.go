package meli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const ListingBatchSize = 5

var (
	ErrListingAuthorizationChanged = errors.New("listing authorization snapshot changed")
	ErrListingReviewChanged        = errors.New("listing review changed or cannot be skipped")
	ErrListingCursor               = errors.New("listing cursor changed or completed")
	ErrListingIdentity             = errors.New("marketplace listing identity mismatch")
	ErrListingTransport            = errors.New("marketplace listing transport failed")
	ErrListingBusy                 = errors.New("listing sync is already running")
	ErrListingTokenExpired         = errors.New("store Access Token expired")
	ErrListingResponse             = errors.New("unexpected marketplace listing response")
	listingID                      = regexp.MustCompile(`^[A-Z]{3}[0-9]{1,30}$`)
	sellerID                       = regexp.MustCompile(`^[0-9]{1,30}$`)
)

type ListingPlatformError int

func (e ListingPlatformError) Error() string {
	return fmt.Sprintf("marketplace listing request returned HTTP %d", e)
}

type ListingScope struct {
	Site     string
	Logistic string
}

type ListingReview struct {
	ListingID string `json:"listing_id,omitempty"`
	Offset    int    `json:"offset"`
	Reason    string `json:"reason"`
	Skippable bool   `json:"skippable"`
}

type listingCursor struct {
	Offset int
	Review ListingReview
}

type ListingPage struct {
	Review       *ListingReview    `json:"review,omitempty"`
	Items        []json.RawMessage `json:"items"`
	HasMore      bool              `json:"has_more"`
	SyncedAt     *time.Time        `json:"synced_at"`
	NextOffset   *int              `json:"next_offset"`
	ScanRequired bool              `json:"scan_required"`
}

type ListingSync struct {
	Review        *ListingReview       `json:"review,omitempty"`
	ConfirmedSkip *ListingReview       `json:"confirmed_skip,omitempty"`
	Saved         int                  `json:"saved"`
	Processed     int                  `json:"processed"`
	Skipped       int                  `json:"skipped"`
	NeedsReview   int                  `json:"needs_review"`
	Partial       int                  `json:"partial"`
	NextOffset    *int                 `json:"next_offset"`
	ScanRequired  bool                 `json:"scan_required"`
	Interruption  *ListingInterruption `json:"interruption,omitempty"`
}

func (s *Service) listingScope(ctx context.Context, owner, store string, scope ListingScope) (tokenSnapshot, string, error) {
	row, err := s.snapshot(ctx, owner, store)
	if err != nil {
		return row, "", err
	}
	if err := tokenStatusError(row.status); err != nil {
		return row, "", err
	}
	if scope.Site == "CBT" && scope.Logistic == "" {
		return row, row.sellerID, nil
	}
	if row.marketsStatus == "verified" && scope.Logistic != "" {
		var markets []Marketplace
		if err := json.Unmarshal(row.marketsJSON, &markets); err != nil {
			return row, "", err
		}
		for _, market := range markets {
			if market.SiteID == scope.Site && market.LogisticType == scope.Logistic {
				return row, market.UserID.String(), nil
			}
		}
	}
	return row, "", ErrSiteNotAuthorized
}

const ListingPageSize = 20

func (s *Service) ListPublished(ctx context.Context, owner, store string, scope ListingScope, offset int) (ListingPage, error) {
	row, seller, err := s.listingScope(ctx, owner, store, scope)
	if err != nil {
		return ListingPage{}, err
	}
	page := ListingPage{Items: []json.RawMessage{}}
	// Recheck the connection in the same SQL statement that reads its private rows.
	rows, err := s.db.Query(ctx, `SELECT sync.synced_at, CASE WHEN sync.store_id IS NULL THEN 0 ELSE sync.next_offset END, COALESCE(sync.scan_required,false), sync.review, item.details
      FROM app.meli_store_connections c
      LEFT JOIN app.meli_listing_syncs sync ON (sync.store_id,sync.site_id,sync.logistic_type,sync.seller_id)=(c.id,$3,$4,$9)
      LEFT JOIN LATERAL (
        SELECT details || jsonb_strip_nulls(jsonb_build_object('detail_unavailable_at',detail_unavailable_at)) AS details, listing_id
        FROM app.meli_published_listings
        WHERE (store_id,site_id,logistic_type)=(c.id,$3,$4)
          AND details->>'seller_id'=$9
        ORDER BY listing_id OFFSET $5 LIMIT $6
      ) item ON true
      WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid AND c.status='connected'
        AND c.marketplaces=$7::jsonb AND c.marketplaces_status=$8
      ORDER BY item.listing_id`, store, owner, scope.Site, scope.Logistic, offset, ListingPageSize+1, row.marketsJSON, row.marketsStatus, seller)
	if err != nil {
		return ListingPage{}, err
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var detail json.RawMessage
		if err = rows.Scan(&page.SyncedAt, &page.NextOffset, &page.ScanRequired, &page.Review, &detail); err != nil {
			return ListingPage{}, err
		}
		found = true
		if detail != nil {
			page.Items = append(page.Items, detail)
		}
	}
	if err = rows.Err(); err != nil {
		return ListingPage{}, err
	}
	if !found {
		return ListingPage{}, ErrListingAuthorizationChanged
	}
	if len(page.Items) > ListingPageSize {
		page.HasMore = true
		page.Items = page.Items[:ListingPageSize]
	}
	return page, nil
}

func (s *Service) SyncPublished(ctx context.Context, owner, store string, scope ListingScope, expected *int, restart bool) (ListingSync, error) {
	// Reserve time for authorization and the final commit before the proxy's upstream idle timeout.
	return s.syncPublished(ctx, owner, store, scope, expected, restart, "", 10*time.Second)
}

func (s *Service) SkipPublishedReview(ctx context.Context, owner, store string, scope ListingScope, offset int, id string) (ListingSync, error) {
	if !listingID.MatchString(id) {
		return ListingSync{}, ErrListingReviewChanged
	}
	return s.syncPublished(ctx, owner, store, scope, &offset, false, id, 10*time.Second)
}

func (s *Service) syncPublished(ctx context.Context, owner, store string, scope ListingScope, expected *int, restart bool, skipID string, budget time.Duration) (ListingSync, error) {
	trace := slog.With("listing_sync_attempt", rand.Text(), "restart", restart)
	readCtx, endRead := context.WithTimeout(ctx, 2*time.Second)
	defer endRead()
	started := time.Now()
	row, seller, err := s.listingScope(readCtx, owner, store, scope)
	logListingStage(trace, "read_authorization", started, err)
	if err != nil {
		return ListingSync{}, err
	}
	// Use the current token without starting the Issue #21 refresh flow.
	if row.expires == nil || !row.expires.After(time.Now().Add(budget)) {
		return ListingSync{}, ErrListingTokenExpired
	}
	auth, err := s.accessFromSnapshot(owner, row)
	if err != nil {
		return ListingSync{}, err
	}
	if !sellerID.MatchString(seller) {
		return ListingSync{}, ErrListingIdentity
	}
	key := strings.ToLower(store)
	if _, busy := s.listingRuns.LoadOrStore(key, struct{}{}); busy {
		return ListingSync{}, ErrListingBusy
	}
	defer s.listingRuns.Delete(key)
	started = time.Now()
	cursor, err := readListingCursor(readCtx, s.db, store, scope, seller)
	endRead()
	logListingStage(trace, "read_cursor", started, err)
	if err != nil {
		return ListingSync{}, err
	}
	offset := cursor.Offset
	if restart {
		offset = 0
	}
	if offset < 0 || (expected != nil && *expected != offset) {
		return ListingSync{}, ErrListingCursor
	}
	if skipID != "" && (!cursor.Review.Skippable || cursor.Review.ListingID != skipID || cursor.Review.Offset != offset) {
		return ListingSync{}, ErrListingReviewChanged
	}
	fetchCtx, endFetch := context.WithTimeout(ctx, budget)
	started = time.Now()
	result, details, err := s.fetchListingBatch(fetchCtx, auth.AccessToken, seller, scope, offset, skipID, trace)
	trace.Info("listing sync stage", "stage", "platform_read_batch", "elapsed_ms", time.Since(started).Milliseconds(),
		"processed", result.Processed, "ready_to_save", result.Saved,
		"platform_budget_exhausted", errors.Is(fetchCtx.Err(), context.DeadlineExceeded))
	endFetch()
	if ctx.Err() != nil {
		return ListingSync{}, ctx.Err()
	}
	if err != nil {
		if errors.Is(err, ErrListingReviewChanged) {
			return ListingSync{}, err
		}
		interruption := listingInterruption(err)
		result = ListingSync{Interruption: interruption}
		switch interruption.Code {
		case "listing_invalid_response", "listing_identity_mismatch":
			reason := "search_response_invalid"
			if interruption.Code == "listing_identity_mismatch" {
				reason = "search_identity_mismatch"
			}
			result.Review = &ListingReview{Offset: offset, Reason: reason}
			result.NeedsReview = 1
		}
		// A failed first search has no progress to replace the persisted restart state.
		if restart {
			return result, interruption
		}
		if result.Review == nil {
			return ListingSync{}, interruption
		}
		result.NextOffset = &offset
	}
	if result.Review == nil && result.Interruption != nil && !restart &&
		cursor.Review.Reason != "" && result.NextOffset != nil && cursor.Review.Offset == *result.NextOffset {
		result.Review = &cursor.Review
	}
	// The platform budget can expire while the original request is still live.
	writeCtx, endWrite := context.WithTimeout(ctx, 3*time.Second)
	defer endWrite()
	if err = s.saveListingBatch(writeCtx, owner, store, scope, seller, row, cursor, result, details, trace); err != nil {
		return ListingSync{}, err
	}
	if result.Saved == 0 && result.Interruption != nil {
		return result, result.Interruption
	}
	return result, nil
}

func readListingCursor(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, store string, scope ListingScope, seller string) (listingCursor, error) {
	var cursor listingCursor
	err := db.QueryRow(ctx, `SELECT CASE WHEN state.store_id IS NULL THEN 0 ELSE COALESCE(state.next_offset,-1) END,
        COALESCE(state.review,'{}'::jsonb)
        FROM (SELECT 1) AS anchor
        LEFT JOIN app.meli_listing_syncs state ON
          (state.store_id,state.site_id,state.logistic_type,state.seller_id)=($1::uuid,$2,$3,$4)`,
		store, scope.Site, scope.Logistic, seller).Scan(&cursor.Offset, &cursor.Review)
	return cursor, err
}

type listingWrite struct {
	detail        json.RawMessage
	partial       bool
	unavailableAt *time.Time
}

func (s *Service) fetchListingBatch(ctx context.Context, token, seller string, scope ListingScope, offset int, skipID string, trace *slog.Logger) (ListingSync, map[string]listingWrite, error) {
	var search struct {
		Seller json.Number `json:"seller_id"`
		IDs    []string    `json:"results"`
		Paging struct {
			Total  *int `json:"total"`
			Offset *int `json:"offset"`
			Limit  *int `json:"limit"`
		} `json:"paging"`
	}
	limit := min(ListingBatchSize, 1000-offset)
	query := url.Values{"offset": {strconv.Itoa(offset)}, "limit": {strconv.Itoa(limit)}, "orders": {"start_time_desc"}}
	started := time.Now()
	var timing listingGETTiming
	_, err := s.listingGET(ctx, token, "/marketplace/users/"+seller+"/items/search?"+query.Encode(), &search, &timing)
	p := search.Paging
	if err == nil {
		switch {
		case search.Seller.String() != seller:
			err = ErrListingIdentity
		case search.IDs == nil || p.Total == nil || *p.Total < 0 || p.Offset == nil || *p.Offset != offset ||
			p.Limit == nil || *p.Limit != limit || len(search.IDs) > limit || (len(search.IDs) == 0 && offset < *p.Total):
			err = ErrListingResponse
		}
	}
	var pagingTotal any
	if p.Total != nil {
		pagingTotal = *p.Total
	}
	logListingStage(trace.With(timing.attrs(ctx)...).With("result_count", len(search.IDs), "paging_total", pagingTotal), "platform_search_get", started, err)
	if err != nil {
		return ListingSync{}, nil, err
	}

	if skipID != "" && (len(search.IDs) == 0 || search.IDs[0] != skipID) {
		return ListingSync{}, nil, ErrListingReviewChanged
	}

	var result ListingSync
	details := make(map[string]listingWrite, len(search.IDs))
	for _, id := range search.IDs {
		if !listingID.MatchString(id) {
			result.Skipped++
			result.Processed++
			continue
		}
		started = time.Now()
		select {
		case <-ctx.Done():
			result.Interruption = listingInterruption(ctx.Err())
			logListingStage(trace.With("processed_before", result.Processed,
				"platform_budget_exhausted", errors.Is(ctx.Err(), context.DeadlineExceeded)), "platform_pacing", started, ctx.Err())
		case <-time.After(250 * time.Millisecond):
		}
		if result.Interruption != nil {
			break
		}
		started = time.Now()
		timing = listingGETTiming{}
		item, reason, fetchErr := s.fetchListingDetail(ctx, token, seller, scope, id, &timing)
		logListingStage(trace.With(timing.attrs(ctx)...).With("processed_before", result.Processed),
			"platform_detail_get", started, fetchErr)
		if errors.Is(fetchErr, ListingPlatformError(http.StatusNotFound)) {
			checkedAt := time.Now()
			details[id] = listingWrite{unavailableAt: &checkedAt}
			result.Skipped++
			result.Processed++
			continue
		}
		if fetchErr != nil {
			if reason != "" {
				review := &ListingReview{ListingID: id, Offset: offset + result.Processed,
					Reason: reason, Skippable: reason == "detail_field_format"}
				if skipID == id && result.Processed == 0 && review.Skippable {
					result.ConfirmedSkip = review
					result.Skipped++
					result.Processed++
					continue
				}
				result.Review = review
				result.NeedsReview++
			}
			result.Interruption = listingInterruption(fetchErr)
			break
		}
		detail, err := json.Marshal(item)
		if err != nil {
			return ListingSync{}, nil, err
		}
		details[id] = listingWrite{detail: detail, partial: item.Partial}
		result.Processed++
	}
	for _, detail := range details {
		if detail.unavailableAt == nil {
			result.Saved++
		}
		if detail.partial {
			result.Partial++
		}
	}
	next := offset + result.Processed
	if result.Interruption != nil || next < *p.Total {
		if next < 1000 {
			result.NextOffset = &next
		} else {
			result.ScanRequired = true
		}
	}
	return result, details, nil
}

func (s *Service) saveListingBatch(ctx context.Context, owner, store string, scope ListingScope, seller string, row tokenSnapshot, cursor listingCursor, result ListingSync, details map[string]listingWrite, trace *slog.Logger) error {
	type timing struct {
		stage, outcome, cause string
		elapsedMS             int64
		attrs                 []slog.Attr
	}
	timings := make([]timing, 0, 8+ListingBatchSize)
	record := func(stage string, started time.Time, err error, attrs ...slog.Attr) {
		timings = append(timings, timing{stage: stage, elapsedMS: time.Since(started).Milliseconds(), outcome: listingStageOutcome(err), attrs: attrs})
	}
	// Emit after rollback and connection release so logging cannot consume the transaction budget.
	defer func() {
		for _, step := range timings {
			level := slog.LevelInfo
			attrs := []slog.Attr{slog.String("stage", step.stage), slog.Int64("elapsed_ms", step.elapsedMS), slog.String("outcome", step.outcome)}
			if step.cause != "" {
				level = slog.LevelWarn
				attrs = append(attrs, slog.String("cause", step.cause))
			}
			trace.LogAttrs(ctx, level, "listing sync stage", append(attrs, step.attrs...)...)
		}
	}()

	before := s.db.Stat()
	started := time.Now()
	conn, err := s.db.Acquire(ctx)
	elapsedMS := time.Since(started).Milliseconds()
	after := s.db.Stat()
	// Deltas describe the shared pool. Empty acquire counters include only successful waits.
	timings = append(timings, timing{
		stage: "db_acquire", elapsedMS: elapsedMS, outcome: listingStageOutcome(err),
		attrs: []slog.Attr{
			slog.Int("pool_total", int(after.TotalConns())),
			slog.Int("pool_idle", int(after.IdleConns())),
			slog.Int("pool_in_use", int(after.AcquiredConns())),
			slog.Int64("pool_new_total", after.NewConnsCount()),
			slog.Int64("pool_wait_total", after.EmptyAcquireCount()),
			slog.Int64("pool_canceled_total", after.CanceledAcquireCount()),
			slog.Int64("new_conns_delta", after.NewConnsCount()-before.NewConnsCount()),
			slog.Int64("empty_acquires_delta", after.EmptyAcquireCount()-before.EmptyAcquireCount()),
			slog.Int64("empty_acquire_wait_ms_delta", (after.EmptyAcquireWaitTime() - before.EmptyAcquireWaitTime()).Milliseconds()),
			slog.Int64("canceled_acquires_delta", after.CanceledAcquireCount()-before.CanceledAcquireCount()),
		},
	})
	if err != nil {
		return err
	}
	defer conn.Release()

	started = time.Now()
	tx, err := conn.Begin(ctx)
	record("db_begin", started, err)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // pgx discards the connection if cancellation prevents rollback.

	started = time.Now()
	err = writeListingBatch(ctx, tx, owner, store, scope, seller, row, cursor, result, details, record)
	record("db_write", started, err)
	if err != nil {
		return err
	}
	started = time.Now()
	err = tx.Commit(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxCommitRollback) {
		step := timing{stage: "db_commit", elapsedMS: time.Since(started).Milliseconds(), cause: listingStageOutcome(err)}
		if pgconn.SafeToRetry(err) {
			step.outcome = "commit_not_sent"
			timings = append(timings, step)
			return err
		}
		step.outcome = "commit_unknown"
		timings = append(timings, step)
		return &ListingInterruption{Code: "listing_commit_unknown"}
	}
	record("db_commit", started, err)
	return err
}

func writeListingBatch(ctx context.Context, tx pgx.Tx, owner, store string, scope ListingScope, seller string, row tokenSnapshot, cursor listingCursor, result ListingSync, details map[string]listingWrite, record func(string, time.Time, error, ...slog.Attr)) error {
	var valid bool
	started := time.Now()
	err := tx.QueryRow(ctx, `SELECT COALESCE(status='connected' AND seller_id=$3
        AND access_token_ciphertext=$4 AND marketplaces=$5::jsonb AND marketplaces_status=$6
        AND access_token_expires_at > clock_timestamp(),false)
        FROM app.meli_store_connections
        WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR UPDATE`,
		store, owner, row.sellerID, row.accessCipher, row.marketsJSON, row.marketsStatus).Scan(&valid)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		err = ErrStoreNotFound
	case err == nil && !valid:
		err = ErrListingAuthorizationChanged
	}
	record("db_authorization_lock", started, err)
	if err != nil {
		return err
	}
	started = time.Now()
	current, err := readListingCursor(ctx, tx, store, scope, seller)
	if err == nil && current != cursor {
		err = ErrListingCursor
	}
	record("db_cursor_recheck", started, err)
	if err != nil {
		return err
	}
	var batch pgx.Batch
	for _, id := range slices.Sorted(maps.Keys(details)) {
		item := details[id]
		if item.unavailableAt != nil {
			started = time.Now()
			_, err = tx.Exec(ctx, `UPDATE app.meli_published_listings SET detail_unavailable_at=$6
              WHERE (store_id,site_id,logistic_type,listing_id)=($1::uuid,$2,$3,$4)
                AND details->>'seller_id'=$5`, store, scope.Site, scope.Logistic, id, seller, item.unavailableAt)
			record("db_mark_unavailable", started, err)
			if err != nil {
				return err
			}
			continue
		}
		detail := item.detail
		if item.partial {
			var previous json.RawMessage
			started = time.Now()
			err = tx.QueryRow(ctx, `SELECT details FROM app.meli_published_listings
              WHERE (store_id,site_id,logistic_type,listing_id)=($1::uuid,$2,$3,$4)
                AND details->>'seller_id'=$5`, store, scope.Site, scope.Logistic, id, seller).Scan(&previous)
			timingErr := err
			if errors.Is(err, pgx.ErrNoRows) {
				timingErr = nil // A missing snapshot is expected for the first partial read.
			}
			record("db_partial_snapshot_read", started, timingErr)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			if err == nil {
				detail, err = mergePartialListing(previous, detail)
				if err != nil {
					return err
				}
			}
		}
		batch.Queue(`INSERT INTO app.meli_published_listings
            (store_id,site_id,logistic_type,listing_id,details) VALUES ($1::uuid,$2,$3,$4,$5)
            ON CONFLICT (store_id,site_id,logistic_type,listing_id)
            DO UPDATE SET details=EXCLUDED.details, detail_unavailable_at=NULL`, store, scope.Site, scope.Logistic, id, detail)
	}
	if batch.Len() > 0 {
		started = time.Now()
		results := tx.SendBatch(ctx, &batch)
		for range batch.Len() {
			if _, err = results.Exec(); err != nil {
				break
			}
		}
		if closeErr := results.Close(); err == nil {
			err = closeErr
		}
		record("db_listing_upsert_batch", started, err, slog.Int("batch_statements", batch.Len()))
		if err != nil {
			return err
		}
	}
	started = time.Now()
	_, err = tx.Exec(ctx, `INSERT INTO app.meli_listing_syncs AS state
      (store_id,site_id,logistic_type,seller_id,synced_at,next_offset,scan_required,review,last_skipped_review)
      VALUES ($1::uuid,$2,$3,$4,CASE WHEN $5 THEN now() END,$6,$7,$8,
        NULLIF($9::jsonb,'null'::jsonb) || jsonb_build_object('confirmed_at',now()))
      ON CONFLICT (store_id,site_id,logistic_type) DO UPDATE SET
        seller_id=EXCLUDED.seller_id, next_offset=EXCLUDED.next_offset, scan_required=EXCLUDED.scan_required,
        review=EXCLUDED.review,
        last_skipped_review=COALESCE(EXCLUDED.last_skipped_review,
          CASE WHEN state.seller_id=EXCLUDED.seller_id THEN state.last_skipped_review END),
        synced_at=COALESCE(EXCLUDED.synced_at, CASE WHEN state.seller_id=EXCLUDED.seller_id THEN state.synced_at END)`,
		store, scope.Site, scope.Logistic, seller, result.Saved > 0 || result.Interruption == nil, result.NextOffset, result.ScanRequired, result.Review, result.ConfirmedSkip)
	record("db_cursor_upsert", started, err)
	if err != nil {
		return err
	}
	return ctx.Err()
}

func logListingStage(trace *slog.Logger, stage string, started time.Time, err error) {
	var platform ListingPlatformError
	errors.As(err, &platform)
	trace.Info("listing sync stage", "stage", stage, "elapsed_ms", time.Since(started).Milliseconds(),
		"outcome", listingStageOutcome(err), "platform_status", int(platform))
}

func listingStageOutcome(err error) string {
	var platform ListingPlatformError
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded), pgconn.Timeout(err):
		return "timeout"
	case errors.Is(err, pgx.ErrTxCommitRollback):
		return "rolled_back"
	case errors.Is(err, ErrListingTransport):
		return "platform_transport"
	case errors.Is(err, ErrListingResponse):
		return "invalid_platform_response"
	case errors.Is(err, ErrListingIdentity):
		return "identity_mismatch"
	case errors.As(err, &platform):
		return "platform_http_error"
	case errors.Is(err, ErrStoreNotFound):
		return "store_not_found"
	case errors.Is(err, ErrSiteNotAuthorized), errors.Is(err, ErrListingAuthorizationChanged):
		return "authorization_rejected"
	case errors.Is(err, ErrListingCursor):
		return "cursor_changed"
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return "connection_eof"
	default:
		return "error"
	}
}

type ListingInterruption struct {
	Code           string `json:"code"`
	PlatformStatus int    `json:"platform_status,omitempty"`
	Retryable      bool   `json:"retryable"`
}

func (e *ListingInterruption) Error() string { return e.Code }

func listingInterruption(err error) *ListingInterruption {
	var remote ListingPlatformError
	switch {
	case errors.Is(err, ErrListingIdentity):
		return &ListingInterruption{Code: "listing_identity_mismatch"}
	case errors.Is(err, context.DeadlineExceeded):
		return &ListingInterruption{Code: "listing_timeout", Retryable: true}
	case errors.Is(err, ErrListingTransport):
		return &ListingInterruption{Code: "listing_transport", Retryable: true}
	case errors.As(err, &remote):
		e := &ListingInterruption{Code: "listing_platform_error", PlatformStatus: int(remote)}
		switch {
		case remote == 401 || remote == 403:
			e.Code = "listing_platform_auth"
		case remote == 429:
			e.Code, e.Retryable = "listing_rate_limited", true
		case remote >= 500:
			e.Retryable = true
		}
		return e
	default:
		return &ListingInterruption{Code: "listing_invalid_response"}
	}
}

func mergePartialListing(previous, incoming json.RawMessage) (json.RawMessage, error) {
	var old, patch map[string]any
	for _, doc := range []struct {
		raw    json.RawMessage
		fields *map[string]any
	}{{previous, &old}, {incoming, &patch}} {
		decoder := json.NewDecoder(bytes.NewReader(doc.raw))
		decoder.UseNumber()
		if err := decoder.Decode(doc.fields); err != nil {
			return nil, err
		}
	}
	return json.Marshal(mergeListingFields(old, patch))
}

func mergeListingFields(old, patch map[string]any) map[string]any {
	for key, value := range patch {
		if value == nil {
			continue
		}
		switch next := value.(type) {
		case map[string]any:
			if previous, ok := old[key].(map[string]any); ok {
				value = mergeListingFields(previous, next)
			}
		case []any:
			// Empty arrays are explicit. Nonempty collections merge by ID.
			switch key {
			case "pictures", "attributes", "variations", "attribute_combinations":
				if previous, ok := old[key].([]any); ok && len(next) > 0 {
					value = mergeListingEntries(previous, next)
				}
			}
		}
		old[key] = value
	}
	return old
}

func mergeListingEntries(previous, next []any) []any {
	indexes := make(map[string]int, len(previous))
	for i, entry := range previous {
		if id, ok := listingEntryID(entry); ok {
			indexes[id] = i
		}
	}
	for _, entry := range next {
		id, ok := listingEntryID(entry)
		if !ok {
			// A partial entry without a stable key cannot safely replace another.
			continue
		}
		if index, exists := indexes[id]; exists {
			previous[index] = mergeListingFields(previous[index].(map[string]any), entry.(map[string]any))
		} else {
			indexes[id] = len(previous)
			previous = append(previous, entry)
		}
	}
	return previous
}

func listingEntryID(entry any) (string, bool) {
	fields, ok := entry.(map[string]any)
	if !ok {
		return "", false
	}
	switch id := fields["id"].(type) {
	case string:
		return id, id != ""
	case json.Number:
		return id.String(), id != ""
	default:
		return "", false
	}
}

func (s *Service) fetchListingDetail(ctx context.Context, token, seller string, scope ListingScope, id string, timing *listingGETTiming) (listingDetail, string, error) {
	var body json.RawMessage
	status, err := s.listingGET(ctx, token, "/marketplace/items/"+id+"?include_attributes=all", &body, timing)
	if errors.Is(err, ErrListingResponse) {
		return listingDetail{}, "detail_payload_unreadable", err
	}
	if err != nil {
		return listingDetail{}, "", err
	}
	started := time.Now()
	defer func() { timing.decode += time.Since(started) }()
	// Verify identity independently of the rest of the payload before classifying a skippable error.
	var identity struct {
		ID       string      `json:"id"`
		Site     string      `json:"site_id"`
		Seller   json.Number `json:"seller_id"`
		Logistic string      `json:"user_logistic_type"`
	}
	if json.Unmarshal(body, &identity) != nil || identity.ID != id || identity.Site != scope.Site ||
		identity.Seller.String() != seller || (scope.Logistic != "" && identity.Logistic != "" && identity.Logistic != scope.Logistic) {
		return listingDetail{}, "detail_identity_mismatch", ErrListingIdentity
	}
	var item listingDetail
	if json.Unmarshal(body, &item) != nil || !validDetailFieldShapes(item) {
		return listingDetail{}, "detail_field_format", ErrListingResponse
	}
	item.Partial = status == http.StatusPartialContent
	return item, "", nil
}

func validDetailFieldShapes(item listingDetail) bool {
	fields := []struct {
		raw   json.RawMessage
		first string
	}{
		{item.Price, "-0123456789"}, {item.Stock, "-0123456789"},
		{item.Currency, "\""}, {item.SKU, "\""}, {item.Pictures, "["},
		{item.Attributes, "["}, {item.Variations, "["}, {item.Markets, "["},
	}
	for _, field := range fields {
		value := bytes.TrimSpace(field.raw)
		if len(value) > 0 && !bytes.Equal(value, []byte("null")) && !strings.ContainsRune(field.first, rune(value[0])) {
			return false
		}
	}
	return true
}

type listingDetail struct {
	Partial        bool            `json:"partial"`
	SitelessID     *string         `json:"siteless_user_product_id,omitempty"`
	SitelessFamily *json.Number    `json:"siteless_family_id,omitempty"`
	ID             string          `json:"id"`
	Site           string          `json:"site_id"`
	Seller         json.Number     `json:"seller_id"`
	Logistic       string          `json:"user_logistic_type,omitempty"`
	Global         string          `json:"cbt_item_id,omitempty"`
	Parent         string          `json:"parent_item_id,omitempty"`
	UserProduct    *string         `json:"user_product_id,omitempty"`
	Family         *json.Number    `json:"family_id,omitempty"`
	FamilyName     *string         `json:"family_name,omitempty"`
	Title          *string         `json:"title,omitempty"`
	Category       *string         `json:"category_id,omitempty"`
	Price          json.RawMessage `json:"price,omitempty"`
	Currency       json.RawMessage `json:"currency_id,omitempty"`
	Stock          json.RawMessage `json:"available_quantity,omitempty"`
	Status         *string         `json:"status,omitempty"`
	LastUpdated    *string         `json:"last_updated,omitempty"`
	Thumbnail      *string         `json:"secure_thumbnail,omitempty"`
	SKU            json.RawMessage `json:"seller_custom_field,omitempty"`
	Pictures       json.RawMessage `json:"pictures,omitempty"`
	Attributes     json.RawMessage `json:"attributes,omitempty"`
	Variations     json.RawMessage `json:"variations,omitempty"`
	Markets        json.RawMessage `json:"marketplace_items,omitempty"`
}

type listingGETTiming struct {
	request, body, decode time.Duration
	failedPhase           string
	retryAfter            string
}

func (timing listingGETTiming) attrs(ctx context.Context) []any {
	deadline, _ := ctx.Deadline()
	remaining := time.Until(deadline)
	return []any{
		"request_to_headers_or_error_ms", timing.request.Milliseconds(), "body_read_ms", timing.body.Milliseconds(),
		"json_decode_ms", timing.decode.Milliseconds(), "failed_phase", timing.failedPhase,
		"platform_budget_remaining_ms", max(0, remaining.Milliseconds()),
		"platform_budget_exhausted", remaining <= 0,
	}
}

func (s *Service) listingGET(ctx context.Context, token, path string, target any, timing *listingGETTiming) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.apiURL+path, nil)
	if err != nil {
		return 0, ErrListingResponse
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	var written atomic.Bool
	if timing == nil {
		timing = &listingGETTiming{}
	} else {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
			WroteRequest: func(info httptrace.WroteRequestInfo) { written.Store(info.Err == nil) },
		}))
	}
	// Do includes connection setup or reuse and ends at response headers or an error.
	started := time.Now()
	resp, err := s.client.Do(req)
	timing.request = time.Since(started)
	if err != nil {
		timing.failedPhase = "request_transport"
		if written.Load() {
			timing.failedPhase = "response_headers"
		}
		return 0, listingReadError(ctx, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		timing.failedPhase = "http_status"
		if resp.StatusCode == http.StatusTooManyRequests {
			timing.retryAfter = resp.Header.Get("Retry-After")
		}
		return resp.StatusCode, ListingPlatformError(resp.StatusCode)
	}
	const maxBody = 2 << 20
	started = time.Now()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	timing.body = time.Since(started)
	if err != nil {
		timing.failedPhase = "response_body"
		return resp.StatusCode, listingReadError(ctx, err)
	}
	if len(data) > maxBody {
		timing.failedPhase = "response_body"
		return resp.StatusCode, ErrListingResponse
	}
	started = time.Now()
	err = json.Unmarshal(data, target)
	timing.decode = time.Since(started)
	if err != nil {
		timing.failedPhase = "json_decode"
		return resp.StatusCode, ErrListingResponse
	}
	return resp.StatusCode, nil
}

func listingReadError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return ErrListingTransport
}
