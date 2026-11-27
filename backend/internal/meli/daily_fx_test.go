package meli

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDailyFXTruncationAndDate(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want string
	}{
		{`{"date":"2026-11-26","base":"USD","quote":"CNY","rate":6.6899}`, "6.6"},
		{`{"date":"2026-11-23","base":"USD","quote":"CNY","rate":6.999}`, "6.9"},
		{`{"date":"2026-11-26","base":"USD","quote":"CNY","rate":7}`, "7.0"},
		{`{"date":"2026-11-27","base":"USD","quote":"CNY","rate":6.7}`, ""},
		{`{"date":"2026-11-19","base":"USD","quote":"CNY","rate":6.7}`, ""},
		{`{"date":"2026-11-26","base":"CNY","quote":"USD","rate":0.15}`, ""},
		{`{"date":"2026-11-26","base":"USD","quote":"CNY","rate":0}`, ""},
	} {
		got, err := parseDailyFX("2026-11-27", []byte(tc.raw))
		if tc.want == "" {
			if err == nil {
				t.Fatalf("bad quote accepted: %+v", got)
			}
		} else if err != nil || got.LockedRate != tc.want {
			t.Fatalf("rate: %+v %v", got, err)
		}
	}
}

func TestDailyFXFirstAcceptedPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PostgreSQL integration requires TEST_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	const day = "1999-02-03"
	if _, err := db.Exec(ctx, `DELETE FROM app.daily_usd_cny_rates WHERE accounting_day=$1::date`, day); err != nil {
		t.Fatal(err)
	}
	defer db.Exec(ctx, `DELETE FROM app.daily_usd_cny_rates WHERE accounting_day=$1::date`, day)
	first := DailyFXRate{Day: day, ObservationDay: "1999-02-02", RawRate: "6.6899", LockedRate: "6.6", Source: dailyFXSource}
	if err := saveDailyFX(ctx, db, first); err != nil {
		t.Fatal(err)
	}
	first.RawRate, first.LockedRate = "7.123", "7.1"
	if err := saveDailyFX(ctx, db, first); err != nil {
		t.Fatal(err)
	}
	// UTC Feb 2 16:00 is Beijing Feb 3 00:00.
	fx, err := lookupDailyFX(ctx, db, time.Date(1999, 2, 2, 16, 0, 0, 0, time.UTC))
	if err != nil || fx.CNYPerUSD != "6.6" || fx.Source != dailyFXSource || fx.AsOf.Format(dailyFXDateFormat) != "1999-02-02" {
		t.Fatalf("locked daily rate: %+v %v", fx, err)
	}
}

type dailyFXTestTransport func(*http.Request) (*http.Response, error)

func (f dailyFXTestTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestDailyFXBackfillSkipsUnavailableDay(t *testing.T) {
	_, db, _, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
		(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,order_items)
		VALUES($1,'MLB','remote','987654','10001','paid','USD',10,'1999-01-05T12:00:00Z','1999-01-05T12:00:00Z','[]'),
		($1,'MLB','remote','987654','10002','paid','USD',10,'1999-01-06T12:00:00Z','1999-01-06T12:00:00Z','[]')`, store); err != nil {
		t.Fatal(err)
	}
	for _, day := range []string{"1999-01-05", "1999-01-06", "1999-01-09"} {
		if _, err := db.Exec(ctx, `DELETE FROM app.daily_usd_cny_rates WHERE accounting_day=$1::date`, day); err != nil {
			t.Fatal(err)
		}
		defer db.Exec(context.Background(), `DELETE FROM app.daily_usd_cny_rates WHERE accounting_day=$1::date`, day)
	}
	var requests []string
	client := &http.Client{Transport: dailyFXTestTransport(func(r *http.Request) (*http.Response, error) {
		day := r.URL.Query().Get("date")
		requests = append(requests, day)
		status, body := http.StatusServiceUnavailable, ""
		if day == "1999-01-05" {
			status, body = http.StatusOK, `{"date":"1999-01-05","base":"USD","quote":"CNY","rate":6.6899}`
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	now := time.Date(1999, 1, 9, 12, 0, 0, 0, time.UTC)
	cursor := ""
	if err := refreshDailyFX(ctx, db, client, now, &cursor); err == nil || cursor != "1999-01-05" {
		t.Fatalf("first pass: %s %v", cursor, err)
	}
	if err := refreshDailyFX(ctx, db, client, now, &cursor); err == nil || cursor != "1999-01-06" {
		t.Fatalf("second pass: %s %v", cursor, err)
	}
	if got := strings.Join(requests, ","); got != "1999-01-08,1999-01-04,1999-01-08,1999-01-05" {
		t.Fatalf("unexpected request order: %s", got)
	}
	rate, err := lookupDailyFX(ctx, db, time.Date(1999, 1, 6, 12, 0, 0, 0, time.UTC))
	if err != nil || rate.CNYPerUSD != "6.6" {
		t.Fatalf("later history not backfilled: %+v %v", rate, err)
	}
	var failedRows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM app.daily_usd_cny_rates WHERE accounting_day IN ('1999-01-05','1999-01-09')`).Scan(&failedRows); err != nil || failedRows != 0 {
		t.Fatalf("invalid rate saved: %d %v", failedRows, err)
	}
}
