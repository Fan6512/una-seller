import { Fragment, useEffect, useRef, useState } from "react";
import { orderDate, orderMoney, orderUSD, type OrderRecord, type OrderScope } from "./orders";
import { StoreRequestError, storeRequest } from "./stores";
import ExpectedProfitPanel from "./ExpectedProfitPanel";

type BillingPreview = {
  items: Array<{
    order_id: string;
    available: boolean;
    checked_at?: string;
    payments: Array<{ status: string; money_release_status: string }>;
    taxes: Array<{ type: string; original_usd?: string; refunded_usd?: string }>;
    charges: Array<{
      detail_id?: string; type: string; subtype?: string; amount: string; currency: string;
      status?: string; reversal_of?: string; debited_from_operation?: string;
    }>;
    charge_groups?: Array<{ currency: string; type: string; subtype?: string; original_total: string }>;
    shipping?: { buyer_cost_usd?: string; seller_charge_usd?: string; seller_net_cost_usd?: string;
      seller_charge_detail_id?: string };
  }>;
};

type BillingEntry = { value: BillingPreview["items"][number]; checked: number | null };

function checkedAt(value: string | undefined): number | null {
  const time = value ? Date.parse(value) : NaN;
  return Number.isFinite(time) ? time : null;
}

function signedUSD(value: string | undefined, sign = "") {
  if (!value) return "待核实";
  if (!/^-?\d+(\.\d{1,8})?$/.test(value)) return "待核实";
  const negative = value.startsWith("-");
  const [whole, fraction = ""] = (negative ? value.slice(1) : value).split(".");
  const padded = fraction.padEnd(3, "0");
  // Round the displayed value to cents without changing the precise receipt.
  const cents = BigInt(whole) * 100n + BigInt(padded.slice(0, 2)) + (padded[2] >= "5" ? 1n : 0n);
  const amount = "USD " + (cents / 100n).toString() + "." + (cents % 100n).toString().padStart(2, "0");
  return (cents === 0n ? "" : negative ? "-" : sign) + amount;
}

export default function OrderProfitCard({ order, scope, disabled, rateLimit }: {
  order: OrderRecord;
  scope: OrderScope | null;
  disabled: boolean;
  rateLimit: {until:number};
}) {
  const [preview, setPreview] = useState<BillingEntry | null>(null);
  const [cacheError, setCacheError] = useState("");
  const [loadingCache, setLoadingCache] = useState(false);
  const billing = preview?.value;
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [cooldownUntil, setCooldownUntil] = useState(()=>rateLimit.until);
  const [remaining, setRemaining] = useState(()=>Math.max(0,Math.ceil((rateLimit.until-Date.now())/1000)));
  const cooldown = remaining > 0;
  useEffect(() => {
    const remainingSeconds = () => Math.max(0,Math.ceil((cooldownUntil-Date.now())/1000));
    setRemaining(remainingSeconds());
    if (cooldownUntil <= Date.now()) return;
    const timer = window.setInterval(()=>{
      const seconds=remainingSeconds();
      setRemaining(seconds);
      if (!seconds) window.clearInterval(timer);
    },1000);
    return () => window.clearInterval(timer);
  }, [cooldownUntil]);
  const active = useRef<AbortController | null>(null);
  useEffect(() => () => active.current?.abort(), []);
  const cacheRead = useRef<AbortController | null>(null);
  const storeID = scope?.storeId, site = scope?.site, logistic = scope?.logistic;
  const orderID = order.order_id;
  useEffect(() => {
    if (!storeID || !site || !logistic) return;
    const request = new AbortController();
    const signal = AbortSignal.any([request.signal, AbortSignal.timeout(10_000)]);
    cacheRead.current = request;
    setLoadingCache(true);
    setCacheError("");
    setPreview(null);
    const query = new URLSearchParams({ site, logistic_type: logistic });
    void storeRequest<BillingPreview>("/api/v1/stores/" + encodeURIComponent(storeID) +
      "/orders/" + encodeURIComponent(orderID) + "/billing-cache?" + query, "GET", signal, false, undefined,
      () => "无法读取已保存账单").then(response => {
        if (signal.aborted) return;
        if (response.items.length !== 1 || response.items[0].order_id !== orderID) {
          throw new Error("已保存账单与当前订单不一致");
        }
        const value = response.items[0];
        const checked = checkedAt(value.checked_at);
        if (value.available && checked === null) throw new Error("已保存账单缺少有效更新时间");
        if (value.available) setPreview({ value, checked });
      }).catch(problem => {
        if (!request.signal.aborted) setCacheError(signal.aborted ? "读取已保存账单超时" :
          problem instanceof Error ? problem.message : "无法读取已保存账单");
      }).finally(() => {
        if (cacheRead.current === request) {
          cacheRead.current = null;
          setLoadingCache(false);
        }
      });
    return () => { request.abort(); if (cacheRead.current === request) cacheRead.current = null; };
  }, [storeID, site, logistic, orderID]);

  async function readBilling() {
    if (!scope || disabled || cooldown || rateLimit.until>Date.now() || active.current) return;
    cacheRead.current?.abort();
    cacheRead.current = null;
    setLoadingCache(false);
    setCacheError("");
    const request = new AbortController();
    active.current = request;
    const signal = AbortSignal.any([request.signal, AbortSignal.timeout(22_000)]);
    setBusy(true);
    setError("");
    try {
      const query = new URLSearchParams({ site: scope.site, logistic_type: scope.logistic });
      const response = await storeRequest<BillingPreview>(
        "/api/v1/stores/" + encodeURIComponent(scope.storeId) + "/orders/billing-preview?" + query,
        "POST", signal, false, JSON.stringify({ order_ids: [order.order_id] }), (code, status) => {
          if (status === 429) return "平台暂时限流，请稍后手动读取账单";
          if (status === 403) return "账单权限不足，请核对店铺授权";
          if (code === "order_auth_changed") return "店铺映射已变化，请重新打开订单";
          return "无法读取官方账单（HTTP " + status + "）";
        });
      if (signal.aborted) return;
      if (response.items.length !== 1 || response.items[0].order_id !== order.order_id) {
        throw new Error("平台账单与当前订单不一致");
      }
      const value = response.items[0];
      const checked = checkedAt(value.checked_at);
      if (!value.available || checked === null) {
        setPreview(previous => previous?.checked != null ? previous : { value, checked: null });
        setError(value.available ? "本次账单不完整，未保存" : "官方本次未返回账单");
      } else {
        setPreview({ value, checked });
      }
    } catch (problem) {
      if (!request.signal.aborted) {
        if (problem instanceof StoreRequestError && problem.status === 429) {
          const header=problem.retryAfter?.trim()??"";
          const delay=/^\d+$/.test(header)?Number(header)*1000:Date.parse(header)-Date.now();
          const until=Date.now()+(Number.isFinite(delay)&&delay>0?delay:60_000);
          rateLimit.until=Math.max(rateLimit.until,until);
          setCooldownUntil(rateLimit.until);
          setRemaining(Math.max(1,Math.ceil((rateLimit.until-Date.now())/1000)));
        }
        setError(signal.aborted ? "查询账单超时，请稍后手动读取" :
          problem instanceof Error ? problem.message : "查询账单失败");
      }
    } finally {
      if (active.current === request) active.current = null;
      if (!request.signal.aborted) setBusy(false);
    }
  }

  const pendingRelease = billing?.payments.some(payment => payment.money_release_status === "pending");
  const status = !billing ? "待读取官方账单" : !billing.available ? "官方未返回账单" :
    pendingRelease ? "账单有待释放款项" : "账单待对账确认";
  const shipping = billing?.shipping;
  const otherCharges = billing?.charges.filter(charge => charge.currency === "USD" &&
    (!shipping?.seller_charge_detail_id || charge.detail_id !== shipping.seller_charge_detail_id)) ?? [];
  const otherGroups = billing?.charge_groups?.filter(group => group.currency === "USD" &&
    (!shipping?.seller_charge_detail_id || group.type !== "CHARGE" || group.subtype !== "CCSI")) ?? [];
  const hasUnverifiedUSDCharges = Boolean(billing?.charges.some(charge => charge.currency !== "USD") ||
    billing?.charge_groups?.some(group => group.currency !== "USD"));
  // The upstream subtotal includes the shipping debit shown above. Do not
  // silently hide other debits in that same group or subtract rounded values.
  const hasAdditionalShippingCharges = Boolean(shipping?.seller_charge_detail_id && otherCharges.some(charge =>
    charge.type === "CHARGE" && charge.subtype === "CCSI" && charge.currency === "USD"));

  return <section className="orders-detail-block" aria-label="订单账单与利润">
    <div className="orders-detail-block-title">
      <h3>订单账单与利润</h3>
      <button type="button" className="orders-table-action" disabled={busy || cooldown || disabled || !scope}
        onClick={() => void readBilling()}>{busy ? "读取中…" : cooldown ? "限流冷却中" : "读取官方账单"}</button>
    </div>
    {cooldown && <p className="orders-muted" role="status">账单冷却剩余约 {remaining} 秒，不会自动重试。</p>}
    {loadingCache && <p className="orders-muted" role="status">正在读取已保存账单…</p>}
    {cacheError && <p className="orders-alert" role="alert">{cacheError}</p>}
    {error && <p className="orders-alert" role="alert">{error}{preview?.checked != null && "。下方保留上一次成功保存的账单。"}</p>}
    <dl className="orders-detail-fields">
      <dt>商品价格</dt><dd>{orderUSD(order.sales_amount_usd)}</dd>
      {shipping && <>
        <dt>物流运输（本货件净成本）</dt><dd>{signedUSD(shipping.seller_net_cost_usd, "-")}</dd>
        <dt>　买家承担运费</dt><dd>{signedUSD(shipping.buyer_cost_usd, "+")}</dd>
        <dt>　卖家承担运费（账单原额）</dt><dd>{signedUSD(shipping.seller_charge_usd, "-")}</dd>
        <dt>运费说明</dt><dd>以下运费为本货件相关金额；同货件多单时未分摊，不可跨订单累加。同一货件的不同费用口径，不可相加或重复扣减。</dd>
      </>}
      <dt>当前官方总计</dt><dd>待核账</dd>
      <dt>实际结算回款</dt><dd>{pendingRelease ? "待释放，待核账" : "待核账"}</dd>
      <dt>结算核账状态</dt><dd>{status}</dd>
      <dt>采购成本</dt><dd>{order.purchase_cost_cny == null ? "待补录" : orderMoney(order.purchase_cost_cny, "CNY")}</dd>
      <dt>实际利润</dt><dd>待核算</dd>
      {preview?.checked != null && <><dt>账单最近成功保存</dt><dd title="北京时间">{orderDate(new Date(preview.checked).toISOString())}</dd></>}
      {preview && preview.checked === null && billing?.available && <><dt>账单保存状态</dt><dd>本次预览未保存</dd></>}
    </dl>
    <ExpectedProfitPanel order={order} scope={scope} disabled={disabled || busy || loadingCache} billingChecked={preview?.checked ?? null}/>
    {billing?.available && <>
      {billing.payments.length > 0 && <p className="orders-muted">
        官方款项状态：{billing.payments.map(payment =>
          [payment.status === "approved" ? "已批准" : payment.status,
            payment.money_release_status === "pending" ? "待释放" : payment.money_release_status === "released" ? "已释放" : payment.money_release_status].filter(Boolean).join(" / ") || "未知").join("；")}
      </p>}
      <h4>官方费用、退款及冲正明细</h4>
      {!otherCharges.length && !billing.taxes.length && !hasUnverifiedUSDCharges && <p className="orders-muted">平台未返回其他费用或税务明细。</p>}
      {hasUnverifiedUSDCharges && <p className="orders-muted">部分官方费用缺少已核实的美元金额，暂不展示。</p>}
      {otherGroups.length > 0 && <dl className="orders-detail-fields">
        {otherGroups.map(group => <Fragment key={[group.currency, group.type, group.subtype].join("|")}>
          <dt>{[group.type, group.subtype].filter(Boolean).join(" / ")} 小计</dt>
          <dd>{signedUSD(group.original_total)}</dd>
        </Fragment>)}
      </dl>}
      {hasAdditionalShippingCharges && <p className="orders-muted">其他同类物流扣项逐项展示；原小计包含上方已展示的运费，不再显示混合小计。</p>}
      {(otherCharges.length > 0 || billing.taxes.length > 0) && <details open={hasAdditionalShippingCharges || undefined}>
        <summary>查看官方项目</summary>
        <ul>{otherCharges.map((charge, index) => <li key={charge.detail_id ?? "charge-" + index}>
          {[charge.type, charge.subtype, charge.status].filter(Boolean).join(" · ")}
          {" "}{signedUSD(charge.amount)}
          {charge.reversal_of && <small> · 关联费用 {charge.reversal_of}</small>}
        </li>)}
        {billing.taxes.map((tax, index) => <li key={"tax-" + index}>
          税项 {tax.type || "未提供类型"} · 原额 {tax.original_usd ? signedUSD(tax.original_usd) : "未提供"}
          {" · "}退回 {tax.refunded_usd ? signedUSD(tax.refunded_usd) : "未提供"}
        </li>)}</ul>
      </details>}
    </>}
  </section>;
}
