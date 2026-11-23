import { useEffect, useState } from "react";
import { authLifetime } from "./auth";
import { orderRecordScope, type OrderRecord, type OrderScope } from "./orders";
import { storeRequest } from "./stores";

const translationsBySession = new WeakMap<AbortSignal, Map<string, string>>();

export function useOrderTitleTranslations(orders: OrderRecord[], scope: OrderScope | null) {
  const [titles, setTitles] = useState<ReadonlyMap<string, string>>(() =>
    translationsBySession.get(authLifetime()) ?? new Map());

  useEffect(() => {
    const lifetime = authLifetime();
    const cached = translationsBySession.get(lifetime) ?? new Map<string, string>();
    const refs = orders.slice(0, 20).flatMap(order => {
      const target = orderRecordScope(order, scope);
      if (!target || !order.order_items.some(line => line.title && !cached.has(line.title))) return [];
      return [{ store_id: target.storeId, site_id: target.site, logistic_type: target.logistic, order_id: order.order_id }];
    });
    if (refs.length === 0) return;
    const controller = new AbortController();
    const signal = AbortSignal.any([controller.signal, lifetime, AbortSignal.timeout(8_000)]);
    void storeRequest<Record<string, string>>("/api/v1/stores/order-title-translations", "POST",
      signal, false, JSON.stringify(refs)).then(response => {
      if (signal.aborted) return;
      for (const [original, translated] of Object.entries(response)) {
        if (typeof translated === "string" && translated) cached.set(original, translated);
      }
      while (cached.size > 600) {
        const oldest = cached.keys().next();
        if (!oldest.done) cached.delete(oldest.value);
      }
      translationsBySession.set(lifetime, cached);
      setTitles(new Map(cached));
    }).catch(() => { /* Official titles stay visible when translation is unavailable. */ });
    return () => controller.abort();
  }, [orders, scope?.storeId, scope?.site, scope?.logistic]);

  return titles;
}

export default function TranslatedOrderTitle({ title, translated, detail }: {
  title: string;
  translated?: string;
  detail?: boolean;
}) {
  const translatedTitle = translated?.trim() || title;
  if (detail) return <>
    <strong title={title}>{translatedTitle}</strong>
    {translatedTitle !== title && <small>原标题：{title}</small>}
  </>;
  return <span className="orders-product-title" title={title}>{translatedTitle}</span>;
}
