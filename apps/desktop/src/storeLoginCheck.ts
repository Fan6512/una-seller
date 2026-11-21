import type { StoreConnection } from "./stores";

export type StoreHealth = {
  status: "checking" | "connected" | "attention" | "empty";
  message: string;
};

const INITIAL_READ_FAILED = "__initial_read_failed__";
const isActive = (item: StoreConnection) => item.status !== "disconnected";

// A user-initiated disconnect is not an authorization failure. Ordinary list
// reads must not wash away a failed verification for an active store.
export function storeHealthFromList(items: StoreConnection[], failedIds: ReadonlySet<string> = new Set()): StoreHealth {
  const active = items.filter(isActive);
  if (!active.length) return {
    status: "empty",
    message: items.length ? "所有店铺已断开，可在设置中重新授权" : "未添加店铺，点击前往设置",
  };
  if (failedIds.has(INITIAL_READ_FAILED) || active.some((item) => failedIds.has(item.id))) {
    return { status: "attention", message: "部分店铺连接检查未成功，请进入设置重新验证" };
  }
  const needsAttention = active.some((item) =>
    item.status !== "connected" || item.marketplaces_status !== "verified");
  return needsAttention
    ? { status: "attention", message: "有店铺需要重新授权或核实站点，请在设置中处理" }
    : { status: "connected", message: active.length + " 个店铺连接状态已核实" };
}

// Recovery after the initial GET failed: confirming one store does not prove
// that other active stores are valid. Keep their failure markers until checked.
// This is pure so the manual recovery branch is testable without rendering App.
export function resolveManualStoreCheck(
  failedIds: ReadonlySet<string>,
  stores: StoreConnection[],
  checkedId: string,
): Set<string> {
  const next = new Set(failedIds);
  if (next.delete(INITIAL_READ_FAILED)) {
    for (const item of stores) {
      if (item.id !== checkedId && (item.status === "connected" || item.status === "token_expired")) {
        next.add(item.id);
      }
    }
  }
  next.delete(checkedId);
  // A deliberate disconnect is not a pending failure.
  for (const item of stores) if (!isActive(item)) next.delete(item.id);
  return next;
}

// Run once when a user session mounts. POST is routed to the existing Go
// ownership/refresh gate and cannot be retried blindly after an unknown result.
export async function checkStoresOnLogin(
  listStores: (fresh?: boolean) => Promise<StoreConnection[]>,
  checkStore: (id: string, manual?: boolean) => Promise<unknown>,
): Promise<{ health: StoreHealth; failedStoreIds: Set<string> }> {
  const initial = await listStores(true);
  // Terminal states require owner action, not a speculative token refresh.
  const eligible = initial.filter((item) =>
    item.status === "connected" || item.status === "token_expired");
  if (!eligible.length) return { health: storeHealthFromList(initial), failedStoreIds: new Set() };
  const outcomes = await Promise.allSettled(eligible.map((item) => checkStore(item.id, false)));
  const failedStoreIds = new Set(eligible.filter((_, index) => outcomes[index].status === "rejected")
    .map((item) => item.id));
  const latest = await listStores(true);
  return { health: storeHealthFromList(latest, failedStoreIds), failedStoreIds };
}
