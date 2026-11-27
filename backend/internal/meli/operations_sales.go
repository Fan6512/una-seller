package meli

import (
	"context"
	"encoding/json"
	"math/big"
	"time"
)

// OperationsSales is a cached merchandise subtotal, not settlement, paid-at
// revenue, or a guarantee of complete platform coverage.
type OperationsSales struct {
	Day             string  `json:"day"`
	AmountUSD       *string `json:"amount_usd"`
	PaidOrders      int     `json:"paid_orders"`
	IncludedOrders  int     `json:"included_orders"`
	ReviewOrders    int     `json:"review_orders"`
	RefundedOrders  int     `json:"refunded_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
	OtherOrders     int     `json:"other_orders"`
	ScopeAvailable  bool    `json:"scope_available"`
	Truncated       bool    `json:"truncated"`
	ExpectedProfitCNY *string `json:"expected_profit_cny"`
	ProfitOrders int `json:"profit_orders"`
	ProfitRate *string `json:"profit_rate"`
}

const operationsSalesLimit = 10000

func salesDay(now time.Time) (string, time.Time, time.Time) {
	local := now.In(time.FixedZone("Beijing", 8*60*60))
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
	return start.Format("2006-01-02"), start.UTC(), start.AddDate(0, 0, 1).UTC()
}

func salesScopeAvailable(summary OperationsSummary, store, site string) bool {
	if store == "" && summary.Truncated {
		return false
	}
	found := false
	for _, row := range summary.Stores {
		if store != "" && row.StoreID != store {
			continue
		}
		// Missing authorization is never evidence of zero sales.
		if !row.Connected {
			return false
		}
		for _, market := range row.Sites {
			if site == "" || site == market {
				found = true
			}
		}
	}
	return found
}

func (s *OperationsSales) addOrder(status string, raw []byte, total *big.Rat) {
	switch status {
	case "paid":
		s.PaidOrders++
	case "partially_refunded", "refunded":
		// Without reconciled refund amounts we exclude these entirely and
		// disclose the count rather than treating the original sale as net.
		s.RefundedOrders++
		return
	case "cancelled":
		s.CancelledOrders++
		return
	default:
		s.OtherOrders++
		return
	}
	var lines []OrderLine
	if err := json.Unmarshal(raw, &lines); err != nil {
		s.ReviewOrders++
		return
	}
	amount := orderSalesAmountUSD(lines)
	value, ok := new(big.Rat).SetString(amount)
	if !ok {
		s.ReviewOrders++
		return
	}
	s.IncludedOrders++
	total.Add(total, value)
}

func (s *Service) operationsSales(ctx context.Context, owner, store, site string, summary OperationsSummary) (OperationsSales, error) {
	day, from, to := salesDay(summary.CheckedAt)
	out := OperationsSales{Day: day, ScopeAvailable: salesScopeAvailable(summary, store, site)}
	if !out.ScopeAvailable {
		return out, nil
	}
	// EXISTS avoids double-counting duplicate marketplace mappings. Each
	// cached order is identified by its existing full scoped primary key.
	rows, err := s.db.Query(ctx, `SELECT o.status,o.order_items,`+expectedProfitFields+`
	FROM app.meli_orders o JOIN app.meli_store_connections c ON c.id=o.store_id
	`+expectedProfitJoins+`
	WHERE c.owner_user_id=$1::uuid AND c.status='connected' AND c.marketplaces_status='verified'
	  AND ($2='' OR c.id::text=$2) AND ($3='' OR o.site_id=$3)
	  AND o.date_created >= $4 AND o.date_created < $5
	  AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
	    WHERE m.value->>'user_id'=o.seller_id AND m.value->>'site_id'=o.site_id
	      AND m.value->>'logistic_type'=o.logistic_type)
	LIMIT $6`, owner, store, site, from, to, operationsSalesLimit+1)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	total := new(big.Rat)
	profitTotal := new(big.Rat)
	count := 0
	for rows.Next() {
		count++
		if count > operationsSalesLimit {
			out.Truncated = true
			break
		}
		var status string
		var raw []byte
		var input expectedProfitInput
		if err := rows.Scan(append([]any{&status, &raw}, input.scanArgs()...)...); err != nil {
			return out, err
		}
		out.addOrder(status, raw, total)
		result := input.result()
		if result.FX != nil {
			rate := result.FX.CNYPerUSD
			out.ProfitRate = &rate
		}
		if result.Profit != nil {
			value, err := ratOf(result.Profit.ProfitCNY)
			if err != nil {
				return out, err
			}
			out.ProfitOrders++
			profitTotal.Add(profitTotal, value)
		}
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	if !out.Truncated && out.IncludedOrders > 0 {
		amount := fixedPoint(roundRatio(total, 100), 2)
		out.AmountUSD = &amount
	}
	if !out.Truncated && out.ProfitOrders > 0 {
		amount := fixedPoint(roundRatio(profitTotal, 100), 2)
		out.ExpectedProfitCNY = &amount
	}
	return out, nil
}
