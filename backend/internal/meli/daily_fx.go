package meli

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collectors/fx"
	"github.com/jackc/pgx/v5/pgxpool"
)

const dailyFXSource = "frankfurter_v2_ecb"
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

// The collector owns external transport; finance owns the accepted date and truncation.
func fetchDailyFX(ctx context.Context, client *http.Client, day string) (DailyFXRate, error) {
	target, err := time.Parse(dailyFXDateFormat, day)
	if err != nil {
		return DailyFXRate{}, errDailyFXInvalid
	}
	raw, err := fx.Fetch(ctx, client, target.AddDate(0, 0, -1))
	if err != nil {
		return DailyFXRate{}, err
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

type dailyFXPassStats struct {
	targets, saved, failed int
}

func refreshDailyFX(ctx context.Context, db *pgxpool.Pool, client *http.Client, now time.Time, historyAfter *string, stats *dailyFXPassStats) error {
	today, _, _ := salesDay(now)
	// Only public dates leave the server. Bound each pass to today's missing
	// reference and rotating missing cached-order days, not one call per order.
	// Advance even on an unavailable historical quote; wrap at the end.
	rows, err := db.Query(ctx, `WITH needed AS (
		SELECT $1::date AS day
		UNION
		SELECT DISTINCT (date_created AT TIME ZONE 'Asia/Shanghai')::date
		FROM app.meli_orders WHERE date_created IS NOT NULL
		  AND (date_created AT TIME ZONE 'Asia/Shanghai')::date BETWEEN DATE '1999-01-05' AND $1::date
	)
	SELECT n.day::text FROM needed n
	WHERE NOT EXISTS (SELECT 1 FROM app.daily_usd_cny_rates r WHERE r.accounting_day=n.day)
	ORDER BY CASE WHEN n.day=$1::date THEN 0
		WHEN n.day > COALESCE(NULLIF($2,'')::date, DATE '0001-01-01') THEN 1 ELSE 2 END,n.day LIMIT 2`, today, *historyAfter)
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
	stats.targets = len(days)
	var failed error
	for _, day := range days {
		if ctx.Err() != nil {
			return errors.Join(failed, ctx.Err())
		}
		if day != today {
			*historyAfter = day
		}
		rate, err := fetchDailyFX(ctx, client, day)
		if err == nil {
			err = saveDailyFX(ctx, db, rate)
		}
		if err != nil {
			stats.failed++
			failed = errors.Join(failed, err)
		} else {
			stats.saved++
		}
	}
	return failed
}

// RunDailyFX is opt-in until the migration and real provider connectivity are
// verified. It never gates startup, order ingestion or notification ACK.
func RunDailyFX(ctx context.Context, db *pgxpool.Pool, logger *log.Logger) {
	// The cursor is bounded process-local state, reset on restart.
	historyAfter := ""
	fx.Run(ctx, logger, func(pass context.Context, client *http.Client, now time.Time) error {
		stats := dailyFXPassStats{}
		err := refreshDailyFX(pass, db, client, now, &historyAfter, &stats)
		if stats.targets > 0 || err != nil {
			result := "ok"
			if err != nil {
				result = "unavailable"
				if stats.saved > 0 {
					result = "partial"
				}
			}
			// Counts only; never log a provider response or database error.
			logger.Printf("%s pass result=%s targets=%d saved=%d failed=%d deferred=%d",
				fx.ID, result, stats.targets, stats.saved, stats.failed, stats.targets-stats.saved-stats.failed)
		}
		return err
	})
}
