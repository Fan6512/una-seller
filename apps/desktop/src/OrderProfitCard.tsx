import { useEffect, useRef, useState } from "react";
import { orderMoney, type OrderRecord, type OrderScope } from "./orders";
import { storeRequest } from "./stores";

type BillingPreview = {
  items: Array<{
    order_id: string;
    available: boolean;
    payments: Array<{ status: string; money_release_status: string }>;
    taxes: Array<{ type: string; original_usd?: string; refunded_usd?: string }>;
    charges: Array<{
      detail_id?: string; type: string; subtype?: string; amount: string; currency: string;
      description?: string; status?: string; reversal_of?: string; debited_from_operation?: string;
    }>;
    charge_groups?: Array<{ currency: string; type: string; subtype?: string; original_total: string }>;
  }>;
};

export default function OrderProfitCard({ order, scope, disabled }: {
  order: OrderRecord;
  scope: OrderScope | null;
  disabled: boolean;
}) {
  const [billing, setBilling] = useState<BillingPreview["items"][number] | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const active = useRef<AbortController | null>(null);
  useEffect(() => () => active.current?.abort(), []);

  async function readBilling() {
    if (!scope || disabled || active.current) return;
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
      setBilling(response.items[0]);
    } catch (problem) {
      if (!request.signal.aborted) {
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

  return <section className="orders-detail-block" aria-label="订单账单与利润">
    <div className="orders-detail-block-title">
      <h3>订单账单与利润</h3>
      <button type="button" className="orders-table-action" disabled={busy || disabled || !scope}
        onClick={() => void readBilling()}>{busy ? "读取中…" : "读取官方账单"}</button>
    </div>
    {error && <p className="orders-alert" role="alert">{error}{billing && "。下方显示上一次读取的账单。"}</p>}
    <dl className="orders-detail-fields">
      <dt>订单销售金额</dt><dd>{orderMoney(order.total_amount, order.currency_id)}</dd>
      <dt>销售数量</dt><dd>{order.order_items.length === 0 ? "平台未提供" : order.order_items.reduce((total, line) => total + line.quantity, 0) + " 件"}</dd>
      <dt>结算核账状态</dt><dd>{status}</dd>
      <dt>平台净收益（USD）</dt><dd>待核账</dd>
      <dt>1688 采购成本（CNY）</dt>
      <dd>{order.purchase_cost_cny == null ? "采购成本待补录" : orderMoney(order.purchase_cost_cny, "CNY")}</dd>
      <dt>打包及其他实付成本（CNY）</dt><dd>待确认</dd>
      <dt>汇率（1 USD 折合 CNY）</dt><dd>待确认汇率、来源及日期</dd>
      <dt>预计利润（USD）</dt><dd>待核账及成本确认</dd>
      <dt>利润率</dt><dd>待核账</dd>
    </dl>
    {billing?.available && <>
      {billing.payments.length > 0 && <p className="orders-muted">
        官方款项状态：{billing.payments.map(payment =>
          [payment.status, payment.money_release_status].filter(Boolean).join(" / ") || "未知").join("；")}
      </p>}
      <h4>官方费用、退款及冲正明细</h4>
      {!billing.charges.length && !billing.taxes.length && <p className="orders-muted">官方未返回可显示的费用或税务明细。</p>}
      {billing.charge_groups && <dl className="orders-detail-fields">
        {billing.charge_groups.map(group => <div key={[group.currency, group.type, group.subtype].join("|")}>
          <dt>{[group.type, group.subtype].filter(Boolean).join(" / ")} 原币小计</dt>
          <dd>{orderMoney(group.original_total, group.currency)}</dd>
        </div>)}
      </dl>}
      {(billing.charges.length > 0 || billing.taxes.length > 0) && <details>
        <summary>查看官方项目</summary>
        <ul>{billing.charges.map((charge, index) => <li key={charge.detail_id ?? "charge-" + index}>
          {[charge.description, charge.type, charge.subtype, charge.status].filter(Boolean).join(" · ")}
          {" "}{charge.currency ? orderMoney(charge.amount, charge.currency) : charge.amount + "（币种待核实）"}
          {charge.reversal_of && <small> · 关联费用 {charge.reversal_of}</small>}
        </li>)}
        {billing.taxes.map((tax, index) => <li key={"tax-" + index}>
          税项 {tax.type || "未提供类型"} · 原额 {tax.original_usd ? orderMoney(tax.original_usd, "USD") : "未提供"}
          {" · "}退回 {tax.refunded_usd ? orderMoney(tax.refunded_usd, "USD") : "未提供"}
        </li>)}</ul>
      </details>}
    </>}
  </section>;
}
