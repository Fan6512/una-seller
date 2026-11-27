import { useEffect, useRef, useState } from "react";
import { type OrderRecord, type OrderScope } from "./orders";
import { storeRequest } from "./stores";

type ExpectedProfit = {
  day: string; revision: string; can_confirm: boolean; has_confirmation: boolean; reason: string;
  net_usd: string | null; purchase_cny: string | null; extra_cny: string | null;
  net_source?: string; net_risk?: string; can_set_extra_cost?: boolean;
  fx: { rate: string; source: string; as_of: string } | null;
  profit: { profit_cny: string; profit_usd: string; margin_percent: string | null; kind: "expected" | "billing_estimate" } | null;
};
const reasons: Record<string, string> = {
  net_unconfirmed: "预计净回款待核实",
  order_or_billing_unavailable: "当前订单或有效账单不可用；退款、取消等订单不自动核算",
  confirmation_stale: "订单或费用证据已变化，请重新核对净额",
  estimate_stale: "账单估算证据已变化，待重新核验",
  extra_missing: "额外成本待补录（真实为零填0）",
  ready_estimate: "账单估算已核算（非平台确认）",
  purchase_missing: "采购成本待补录",
  fx_missing: "订单日期固定汇率待获取",
  inputs_invalid: "核算输入异常",
  ready: "已核算（非实际结算）",
};

export default function ExpectedProfitPanel({order, scope, disabled, billingChecked}: {
  order: OrderRecord; scope: OrderScope | null; disabled: boolean; billingChecked: number | null;
}) {
  const [data,setData]=useState<ExpectedProfit|null>(null);
  const [net,setNet]=useState("");
  const [extra,setExtra]=useState("");
  const [editing,setEditing]=useState(false);
  const [costOnly,setCostOnly]=useState(false);
  const [busy,setBusy]=useState(false);
  const [error,setError]=useState("");
  const [refresh,setRefresh]=useState(0);
  const action=useRef<AbortController|null>(null);
  const store=scope?.storeId, site=scope?.site, logistic=scope?.logistic;
  const path=store && site && logistic ? "/api/v1/stores/"+encodeURIComponent(store)+"/orders/"+
    encodeURIComponent(order.order_id)+"/expected-profit?"+new URLSearchParams({site,logistic_type:logistic}) : "";
  useEffect(()=>{
    const controller=new AbortController();
    action.current?.abort();
    setBusy(false); setData(null); setEditing(false); setCostOnly(false); setNet(""); setExtra(""); setError("");
    if (!path) return ()=>controller.abort();
    void storeRequest<ExpectedProfit>(path,"GET",AbortSignal.any([controller.signal,AbortSignal.timeout(10000)]),false,undefined,
      ()=>"预计利润暂不可用（可能尚未完成服务端升级）").then(value=>{
      if (!controller.signal.aborted) setData(value);
    }).catch(problem=>{
      if (!controller.signal.aborted) setError(problem instanceof Error?problem.message:"读取预计利润失败");
    });
    return ()=>{controller.abort(); action.current?.abort();};
  },[path,order.status,order.last_updated,order.purchase_cost_cny,billingChecked,refresh]);
  async function save(clear: boolean) {
    if (!path || !data || disabled || action.current) return;
    const onlyCost=costOnly && !clear;
    if (!clear && ((!onlyCost && (!data.can_confirm || !/^-?(0|[1-9][0-9]{0,15})(\.[0-9]{1,8})?$/.test(net))) ||
        (onlyCost && !data.can_set_extra_cost) || !/^(0|[1-9][0-9]{0,11})(\.[0-9]{1,2})?$/.test(extra))) {
      setError(onlyCost?"请输入真实人民币额外成本（实际为零请填0）":"请输入已核实的订单级美元预计净额及人民币额外成本（实际为零请填0）"); return;
    }
    if (!window.confirm(clear ? "清除这笔订单的预计净额确认？额外成本保留。" : onlyCost ?
      "确认这笔订单的人民币额外成本准确？这不会确认平台净额。" :
      "确认卖家中心显示的是当前订单的美元预计净回款（已包含应扣费用与退款，非商品售价或货件合计），且额外成本准确？")) return;
    const request=new AbortController();
    action.current=request; setBusy(true); setError("");
    try {
      const value=await storeRequest<ExpectedProfit>(path,"POST",AbortSignal.any([request.signal,AbortSignal.timeout(10000)]),false,
        JSON.stringify({revision:data.revision,net_usd:clear||onlyCost?"":net,extra_cny:clear?"":extra,clear,confirmed:true,...(onlyCost?{extra_cost_only:true}:{})}),
        (_code,status)=>status===409?"订单、账单或确认记录已变化，请刷新核对":"预计净额保存未确认，请刷新核对");
      if (!request.signal.aborted) {setData(value);setEditing(false);setCostOnly(false);}
    } catch(problem) {
      if (!request.signal.aborted) {setError(problem instanceof Error?problem.message:"保存结果未确认，请刷新核对");setData(null);setEditing(false);}
    } finally {
      if(action.current===request) {action.current=null;setBusy(false);}
    }
  }
  const isEstimate=data?.net_source==="billing_estimate";
  return <div className="orders-detail-block">
    <h4>预计利润（人民币）</h4>
    <dl className="orders-detail-fields">
      <dt>预计净回款（USD）</dt><dd>{data?.net_usd==null?"待核实":"USD "+data.net_usd+(isEstimate?"（账单估算）":"（卖家中心人工确认）")}</dd>
      <dt>订单日固定汇率</dt><dd>{data?.fx?data.fx.rate+" · "+data.day+"（参考价 "+data.fx.as_of.slice(0,10)+"）":"待获取"}</dd>
      <dt>采购成本</dt><dd>{data?.purchase_cny==null?"待补录":"¥"+data.purchase_cny}</dd>
      <dt>额外成本</dt><dd>{data?.extra_cny==null?"待确认":"¥"+data.extra_cny}</dd>
      <dt>{isEstimate?"预计净利润（账单估算）":"预计净利润"}</dt><dd><strong>{data?.profit?"¥"+data.profit.profit_cny:"—"}</strong></dd>
      <dt>预计利润率</dt><dd>{data?.profit?.margin_percent==null?"—":data.profit.margin_percent+"%"}</dd>
    </dl>
    <p className="orders-muted">{data?reasons[data.reason]??"待核实":"预计利润尚未读取"}；不代表实际到账。</p>
    {isEstimate && <p className="orders-muted">{data?.net_risk==="shipment_membership_unknown"?"官方货件成员关系未完整核验；":""}账单估算仅供经营参考，费用完整性仍待核对。</p>}
    {error && <p className="orders-alert" role="alert">{error}</p>}
    <button type="button" className="orders-table-action" disabled={disabled || busy || !path} onClick={()=>setRefresh(n=>n+1)}>刷新预计利润</button>
    {data?.can_confirm && !editing && <button type="button" className="orders-table-action" disabled={disabled || busy}
      onClick={()=>{setCostOnly(false);setNet(isEstimate?"":data.net_usd??"");setExtra(data.extra_cny??"");setEditing(true);}}>核实净额（可选）</button>}
    {data?.can_set_extra_cost && !editing && <button type="button" className="orders-table-action" disabled={disabled || busy}
      onClick={()=>{setCostOnly(true);setNet("");setExtra(data.extra_cny??"");setEditing(true);}}>编辑额外成本</button>}
    {data?.has_confirmation && !editing && <button type="button" className="orders-table-action" disabled={disabled || busy} onClick={()=>void save(true)}>清除确认</button>}
    {editing && (costOnly?data?.can_set_extra_cost:data?.can_confirm) && <div>
      <p className="orders-muted">{costOnly?"仅保存实际额外成本，不确认净额。":"仅在已核对卖家中心同一订单净额后使用。未核实无需填写。"}</p>
      {!costOnly && <label>订单预计净回款 USD <input value={net} inputMode="decimal" disabled={busy||disabled} onChange={e=>setNet(e.target.value)}/></label>}
      <label>额外成本 CNY <input value={extra} inputMode="decimal" placeholder="真实为零填写0" disabled={busy||disabled} onChange={e=>setExtra(e.target.value)}/></label>
      <button type="button" className="orders-table-action" disabled={busy||disabled} onClick={()=>void save(false)}>确认保存</button>
      <button type="button" className="orders-table-action" disabled={busy} onClick={()=>setEditing(false)}>取消</button>
    </div>}
  </div>;
}
