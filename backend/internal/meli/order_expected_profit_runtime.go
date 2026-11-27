package meli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

var profitRevisionPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Both single-order reads and dashboard aggregation use the same inputs.
// No live provider requests in either path.
const expectedProfitEvidence = `encode(sha256(convert_to(jsonb_build_array(
	o.store_id,c.seller_id,o.site_id,o.logistic_type,o.seller_id,o.order_id,o.status,
	o.pack_id,o.shipment_id,extract(epoch from o.date_created),extract(epoch from o.last_updated),
	o.currency_id,o.total_amount,o.order_items,
	o.billing_snapshot->'shipment_id',(o.billing_snapshot->'preview')-'checked_at'
	)::text,'UTF8')),'hex')`
const expectedProfitRevision = `encode(sha256(convert_to(jsonb_build_array(
	` + expectedProfitEvidence + `,o.expected_profit_confirmation)::text,'UTF8')),'hex')`
const expectedProfitEligible = `COALESCE(o.status='paid'
	AND o.billing_snapshot->>'shipment_id'=COALESCE(o.shipment_id,'')
	AND o.billing_snapshot->'preview'->>'order_id'=o.order_id
	AND o.billing_snapshot->'preview'->>'available'='true'
	AND o.billing_snapshot->'preview'->>'checked_at' IS NOT NULL,false)`
const expectedProfitFields = `o.date_created,` + expectedProfitEligible + `,` +
	expectedProfitEvidence + `,` + expectedProfitRevision + `,
	o.expected_profit_confirmation,trim_scale(p.cost_cny)::text,
	f.locked_rate::text,f.source,f.observation_day::text`
const expectedProfitJoins = `LEFT JOIN app.meli_order_purchases p ON
	(p.store_id,p.site_id,p.logistic_type,p.seller_id,p.order_id)=
	(o.store_id,o.site_id,o.logistic_type,o.seller_id,o.order_id)
	LEFT JOIN app.daily_usd_cny_rates f ON f.accounting_day=(o.date_created AT TIME ZONE 'Asia/Shanghai')::date`

// Require exactly one current child identity for single-order operations.
const expectedProfitScope = `c.id=$1::uuid AND c.owner_user_id=$2::uuid
	AND c.status='connected' AND c.marketplaces_status='verified'
	AND (SELECT count(DISTINCT m.value->>'user_id') FROM jsonb_array_elements(c.marketplaces) m(value)
		WHERE m.value->>'site_id'=$3 AND m.value->>'logistic_type'=$4)=1
	AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		WHERE m.value->>'site_id'=o.site_id AND m.value->>'logistic_type'=o.logistic_type
		AND m.value->>'user_id'=o.seller_id)
	AND o.site_id=$3 AND o.logistic_type=$4 AND o.order_id=$5`

type expectedProfitConfirmation struct {
	NetUSD     string    `json:"net_usd"`
	ExtraCNY   string    `json:"extra_cny"`
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
	Evidence   string    `json:"evidence"`
}

type OrderExpectedProfit struct {
	Day             string                  `json:"day"`
	Revision        string                  `json:"revision"`
	CanConfirm      bool                    `json:"can_confirm"`
	HasConfirmation bool                    `json:"has_confirmation"`
	Reason          string                  `json:"reason"`
	NetUSD          *string                 `json:"net_usd"`
	ExtraCNY        *string                 `json:"extra_cny"`
	PurchaseCNY     *string                 `json:"purchase_cny"`
	FX              *ProfitFXRate           `json:"fx"`
	Profit          *ExpectedOrderProfitUSD `json:"profit"`
}

type expectedProfitInput struct {
	created                time.Time
	eligible               bool
	evidence, revision     string
	confirmation           []byte
	purchase, rate, source sql.NullString
	observation            sql.NullString
}

func (i *expectedProfitInput) scanArgs() []any {
	return []any{&i.created, &i.eligible, &i.evidence, &i.revision, &i.confirmation,
		&i.purchase, &i.rate, &i.source, &i.observation}
}

func (i expectedProfitInput) result() OrderExpectedProfit {
	day, _, _ := salesDay(i.created)
	out := OrderExpectedProfit{Day: day, Revision: i.revision, CanConfirm: i.eligible,
		HasConfirmation: len(i.confirmation)>0, Reason: "net_unconfirmed"}
	if i.purchase.Valid && purchaseCostPattern.MatchString(i.purchase.String) {
		out.PurchaseCNY = &i.purchase.String
	}
	if i.rate.Valid && i.source.Valid && i.observation.Valid {
		observed, err := time.Parse(dailyFXDateFormat, i.observation.String)
		if err == nil && i.source.String == dailyFXSource && fxRatePattern.MatchString(i.rate.String) {
			out.FX = &ProfitFXRate{CNYPerUSD: i.rate.String, Source: i.source.String, AsOf: observed}
		}
	}
	if !i.eligible {
		out.Reason = "order_or_billing_unavailable"
		return out
	}
	var c expectedProfitConfirmation
	if len(i.confirmation) == 0 {
		return out
	}
	if json.Unmarshal(i.confirmation, &c) != nil || c.Evidence != i.evidence ||
		c.Source != ExpectedNetSellerCenter || c.ObservedAt.IsZero() ||
		!moneyUSDPattern.MatchString(c.NetUSD) || !purchaseCostPattern.MatchString(c.ExtraCNY) {
		out.Reason = "confirmation_stale"
		return out
	}
	out.NetUSD, out.ExtraCNY = &c.NetUSD, &c.ExtraCNY
	if out.PurchaseCNY == nil {
		out.Reason = "purchase_missing"
		return out
	}
	if out.FX == nil {
		out.Reason = "fx_missing"
		return out
	}
	value, err := EstimateProfitFromExpectedNet(ConfirmedExpectedUSDNet{Amount: c.NetUSD, Source: c.Source, ObservedAt: c.ObservedAt},
		*out.FX, *out.PurchaseCNY, c.ExtraCNY)
	if err != nil {
		out.Reason = "inputs_invalid"
		return out
	}
	out.Profit, out.Reason = &value, "ready"
	return out
}

func validExpectedProfitScope(owner, store string, scope OrderScope, orderID string) bool {
	return validUUID(owner) && validUUID(store) && orderScopeValid(scope) && billingIDsValid([]string{orderID})
}

func (s *Service) GetExpectedOrderProfit(ctx context.Context, owner, store string, scope OrderScope, orderID string) (OrderExpectedProfit, error) {
	if !validExpectedProfitScope(owner, store, scope, orderID) {
		return OrderExpectedProfit{}, ErrOrderBillingInvalid
	}
	var input expectedProfitInput
	err := s.db.QueryRow(ctx, `SELECT `+expectedProfitFields+`
	FROM app.meli_orders o JOIN app.meli_store_connections c ON c.id=o.store_id
	`+expectedProfitJoins+` WHERE `+expectedProfitScope,
		store, owner, scope.Site, scope.Logistic, orderID).Scan(input.scanArgs()...)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrderExpectedProfit{}, ErrOrderNotFound
	}
	if err != nil {
		return OrderExpectedProfit{}, err
	}
	return input.result(), nil
}

// ConfirmExpectedOrderProfit records an explicit owner attestation of the
// order-level USD expected net read in Seller Center, not an official write.
// The CAS revision binds the full current order and previous confirmation.
func (s *Service) ConfirmExpectedOrderProfit(ctx context.Context, owner, store string, scope OrderScope, orderID, revision, netUSD, extraCNY string, clear bool) (OrderExpectedProfit, error) {
	if !validExpectedProfitScope(owner, store, scope, orderID) || !profitRevisionPattern.MatchString(revision) ||
		(clear && (netUSD != "" || extraCNY != "")) ||
		(!clear && (!moneyUSDPattern.MatchString(netUSD) || !purchaseCostPattern.MatchString(extraCNY))) {
		return OrderExpectedProfit{}, ErrOrderBillingInvalid
	}
	// Lock the connection; the UPDATE rechecks evidence and revision.
	var id string
	err := s.db.QueryRow(ctx, `WITH authorized AS (
		SELECT c.* FROM app.meli_store_connections c
		WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid FOR SHARE
	)
	UPDATE app.meli_orders o SET expected_profit_confirmation=CASE WHEN $9::boolean THEN NULL ELSE
		jsonb_build_object('net_usd',$7::text,'extra_cny',$8::text,'source',$10::text,
		'observed_at',to_jsonb(clock_timestamp()),'evidence',`+expectedProfitEvidence+`) END
	FROM authorized c WHERE o.store_id=c.id AND `+expectedProfitScope+`
	AND `+expectedProfitRevision+`=$6 AND ($9::boolean OR `+expectedProfitEligible+`)
	RETURNING o.order_id`, store, owner, scope.Site, scope.Logistic, orderID, revision, netUSD, extraCNY, clear, ExpectedNetSellerCenter).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrderExpectedProfit{}, ErrOrderAuthorizationChanged
	}
	if err != nil {
		return OrderExpectedProfit{}, err
	}
	return s.GetExpectedOrderProfit(ctx, owner, store, scope, orderID)
}
