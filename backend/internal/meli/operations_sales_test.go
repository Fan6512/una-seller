package meli

import (
	"math/big"
	"testing"
	"time"
)

func TestOperationsSalesDayAndAmounts(t *testing.T) {
	now := time.Date(2026, 11, 26, 16, 0, 0, 0, time.UTC)
	day, start, end := salesDay(now)
	if day != "2026-11-27" || !start.Equal(now) || end.Sub(start) != 24*time.Hour {
		t.Fatalf("wrong Beijing window: %s %s %s", day, start, end)
	}
	var s OperationsSales
	total := new(big.Rat)
	valid := []byte(`[{"item_id":"MLB1","quantity":2,"unit_price":"3.125","currency_id":"USD"}]`)
	s.addOrder("paid", valid, total)
	s.addOrder("paid", []byte(`[{"item_id":"MLB2","quantity":1,"unit_price":"20","currency_id":"BRL"}]`), total)
	s.addOrder("partially_refunded", valid, total)
	s.addOrder("refunded", valid, total)
	s.addOrder("cancelled", valid, total)
	s.addOrder("payment_required", valid, total)
	if total.RatString() != "25/4" || s.PaidOrders != 2 || s.IncludedOrders != 1 ||
		s.ReviewOrders != 1 || s.RefundedOrders != 2 || s.CancelledOrders != 1 || s.OtherOrders != 1 {
		t.Fatalf("incorrect subtotal: %s %+v", total, s)
	}
}
