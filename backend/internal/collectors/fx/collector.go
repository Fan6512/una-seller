// Package fx is the built-in collector.fx plugin. It has no financial
// storage, order, credentials or settlement dependencies.
package fx

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

const ID = "collector.fx"
const endpoint = "https://api.frankfurter.dev/v2/rate/USD/CNY"

var errInvalidResponse = errors.New("daily FX provider unavailable")

// Fetch sends only a public observation date to a fixed provider endpoint.
func Fetch(ctx context.Context, client *http.Client, requested time.Time) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?providers=ECB&date="+requested.Format("2006-01-02"), nil)
	if err != nil {
		return nil, errInvalidResponse
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errInvalidResponse
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil || len(raw) > 4096 {
		return nil, errInvalidResponse
	}
	return raw, nil
}

// Run is the FX plugin's bounded lifecycle. The caller supplies the finance
// use case, which chooses missing accounting days and persists accepted rates.
// No general scheduler or dynamic loading is needed for a single collector.
func Run(ctx context.Context, logger *log.Logger, refresh func(context.Context, *http.Client, time.Time) error) {
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
		err := refresh(pass, client, time.Now())
		cancel()
		if err != nil && ctx.Err() == nil {
			logger.Printf("%s refresh unavailable; retry next interval", ID)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
