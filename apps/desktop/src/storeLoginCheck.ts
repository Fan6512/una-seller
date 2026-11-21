import type { StoreConnection } from "./stores";

export type StoreHealth = {
  status: "checking" | "connected" | "attention" | "empty";
  message: string;
};

// A status read is not a fresh verification. Keep a prior failed check visible
// until that specific store has successfully been checked again.
export function storeHealthFromList(items: StoreConnection[], failedIds: ReadonlySet<string> = new Set()): StoreHealth {
  if (!items.length) return { status: "empty", message: "未添加店铺，点击前往设置" };
  if (failedIds.size > 0) return {
    status: "attention", message: "部分店铺连接检查未成功，请进入设置重新验证",
  };
  const needsAttention = items.some((item) =>
    item.status !== "connected" || item.marketplaces_status !== "verified");
  return needsAttention
    ? { status: "attention", message: "有店铺需要重新授权或核实站点，请在设置中处理" }
    : { status: "connected", message: items.length + " 个店铺连接状态已核实" };
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
