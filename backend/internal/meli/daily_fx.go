package meli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const dailyFXSource = "frankfurter_v2_ecb"
const dailyFXEndpoint = "https://api.frankfurter.dev/v2/rate/USD/CNY"
const dailyFXDateFormat = "2006-01-02"

var dailyFXPattern = regexp.MustCompile(`^[0-9]{1,2}(\.[0-9]{1,12})?$`)
var errDailyFXInvalid = errors.New("invalid daily USD/CNY reference quote")

type DailyFXRate struct {
	Day            string
	ObservationDay string
	RawRate        string
	LockedRate     string
	Source         string
}

func parseDailyFX(day string, raw []byte) (DailyFXRate, error) {
	target, err := time.Parse(dailyFXDateFormat, day)
	if err != nil {
		return DailyFXRate{}, errDailyFXInvalid
	}
	var data struct {
		Date  string      `json:"date"`
		Base  string      `json:"base"`
		Quote string      `json:"quote"`
		Rate  json.Number `json:"rate"`
	}
	if json.Unmarshal(raw, &data) != nil || data.Base != "USD" || data.Quote != "CNY" ||
		!dailyFXPattern.MatchString(data.Rate.String()) {
		return DailyFXRate{}, errDailyFXInvalid
	}
	observation, err := time.Parse(dailyFXDateFormat, data.Date)
	if err != nil || !observation.Before(target) || observation.Before(target.AddDate(0, 0, -7)) {
		return DailyFXRate{}, errDailyFXInvalid
	}
	value, ok := new(big.Rat).SetString(data.Rate.String())
	if !ok || value.Cmp(big.NewRat(1, 10)) < 0 || value.Cmp(big.NewRat(100, 1)) >= 0 {
		return DailyFXRate{}, errDailyFXInvalid
	}
	scaled := new(big.Rat).Mul(value, big.NewRat(10, 1))
	tenths := new(big.Int).Quo(scaled.Num(), scaled.Denom())
	return DailyFXRate{
		Day: day, ObservationDay: data.Date, RawRate: data.Rate.String(),
		LockedRate: fixedPoint(tenths.String(), 1), Source: dailyFXSource,
	}, nil
}

// Fixed previous-calendar-day query: a deterministic opening reference for
// today's orders and late historical backfill. Never send order/account data.
func fetchDailyFX(ctx context.Context, client *http.Client, day string) (DailyFXRate, error) {
	target, err := time.Parse(dailyFXDateFormat, day)
	if err != nil {
		return DailyFXRate{}, errDailyFXInvalid
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dailyFXEndpoint+"?providers=ECB&date="+target.AddDate(0, 0, -1).Format(dailyFXDateFormat), nil)
	if err != nil {
		return DailyFXRate{}, errDailyFXInvalid
	}
	resp, err := client.Do(req)
	if err != nil {
		return DailyFXRate{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return DailyFXRate{}, errDailyFXInvalid
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil || len(raw) > 4096 {
		return DailyFXRate{}, errDailyFXInvalid
	}
	return parseDailyFX(day, raw)
}

func saveDailyFX(ctx context.Context, db *pgxpool.Pool, rate DailyFXRate) error {
	_, err := db.Exec(ctx, `INSERT INTO app.daily_usd_cny_rates
		(accounting_day,observation_day,raw_rate,locked_rate,source)
		VALUES($1::date,$2::date,$3::numeric,$4::numeric,$5)
		ON CONFLICT (accounting_day) DO NOTHING`,
		rate.Day, rate.ObservationDay, rate.RawRate, rate.LockedRate, rate.Source)
	return err
}

// lookupDailyFX performs only a cached read, not a provider call. Missing rows
// must remain unavailable; never substitute the latest rate for an old order.
func lookupDailyFX(ctx context.Context, db *pgxpool.Pool, orderCreatedAt time.Time) (ProfitFXRate, error) {
	day, _, _ := salesDay(orderCreatedAt)
	var fx ProfitFXRate
	var observed string
	err := db.QueryRow(ctx, `SELECT locked_rate::text,source,observation_day::text
		FROM app.daily_usd_cny_rates WHERE accounting_day=$1::date`, day).
		Scan(&fx.CNYPerUSD, &fx.Source, &observed)
	if err != nil {
		return ProfitFXRate{}, err
	}
	fx.AsOf, err = time.Parse(dailyFXDateFormat, observed)
	return fx, err
}

func refreshDailyFX(ctx context.Context, db *pgxpool.Pool, client *http.Client, now time.Time) error {
	today, _, _ := salesDay(now)
	// Only public dates leave the server. Bound each pass to today's missing
	// reference and one oldest missing cached-order day, not one call per order.
	rows, err := db.Query(ctx, `WITH needed AS (
		SELECT $1::date AS day
		UNION
		SELECT DISTINCT (date_created AT TIME ZONE 'Asia/Shanghai')::date
		FROM app.meli_orders WHERE date_created IS NOT NULL
		  AND (date_created AT TIME ZONE 'Asia/Shanghai')::date BETWEEN DATE '1999-01-05' AND $1::date
	)
	SELECT n.day::text FROM needed n
	WHERE NOT EXISTS (SELECT 1 FROM app.daily_usd_cny_rates r WHERE r.accounting_day=n.day)
	ORDER BY CASE WHEN n.day=$1::date THEN 0 ELSE 1 END,n.day LIMIT 2`, today)
	if err != nil {
		return err
	}
	var days []string
	for rows.Next() {
		var day string
		if err := rows.Scan(&day); err != nil {
			rows.Close()
			return err
		}
		days = append(days, day)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, day := range days {
		rate, err := fetchDailyFX(ctx, client, day)
		if err != nil {
			return err // Retry next interval, never write a guessed value.
		}
		if err := saveDailyFX(ctx, db, rate); err != nil {
			return err
		}
	}
	return nil
}

// RunDailyFX is opt-in until the migration and real provider connectivity are
// verified. It never gates startup, order ingestion or notification ACK.
func RunDailyFX(ctx context.Context, db *pgxpool.Pool, logger *log.Logger) {
	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		pass, cancel := context.WithTimeout(ctx, 25*time.Second)
		err := refreshDailyFX(pass, db, client, time.Now())
		cancel()
		if err != nil && ctx.Err() == nil {
			logger.Print("daily FX refresh unavailable; retry next interval")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
