import { useState, useSyncExternalStore } from "react";

export const diagnosticsEnabled = import.meta.env.VITE_UNA_SELLER_DEV_DIAGNOSTICS === "1";

type Operation = "stores_get" | "listings_get" | "description_get" | "store_token_check" | "sync_post"
  | "orders_get" | "order_state_get" | "order_detail_get" | "order_purchase_get"
  | "orders_sync_post" | "orders_skip_post" | "orders_reset_post" | "order_purchase_post";
type Phase = "token_preflight" | "response_headers" | "response_body" | "failed";
type Entry = {
  time: string;
  request_id: number;
  operation: Operation;
  phase: Phase;
  duration_ms: number;
  http_status?: number;
  error_code?: string;
  token_cached?: boolean;
  cache_hit?: boolean;
  server_timing?: string;
};

const limit = 200;
const entries: Entry[] = [];
const listeners = new Set<() => void>();
const knownCodes = new Set([
  "description_unavailable", "description_link_missing", "listing_unavailable", "listing_method",
  "listing_store_missing", "listing_site_unverified", "listing_reauthorization", "listing_platform_auth", "listing_platform_error",
  "listing_token_expired", "listing_refresh_unknown", "listing_auth_changed", "listing_cursor_changed",
  "listing_busy", "listing_rate_limited", "listing_timeout", "listing_commit_unknown", "listing_transport",
  "listing_identity_mismatch", "listing_review_changed", "listing_review_invalid", "listing_offset",
  "listing_canceled", "listing_invalid_response",
  "order_purchase_conflict", "order_needs_review", "order_review_conflict",
  "order_invalid_response", "order_platform_unavailable", "order_site_not_authorized",
  "order_reauth_required", "order_refresh_unknown", "order_busy", "order_auth_changed",
  "order_rate_limited", "order_timeout", "order_canceled", "order_reset_conflict",
  "order_reset_invalid", "order_unavailable",
  "order_scope_required", "order_skip_invalid", "order_offset_invalid", "order_status_invalid",
]);
let cursor = 0;
let nextRequestId = 0;
let version = 0;
let lifetime = 0;

function changed() {
  version++;
  for (const listener of listeners) listener();
}

export function clearDiagnostics() {
  lifetime++;
  cursor = 0;
  nextRequestId = 0;
  entries.length = 0;
  changed();
}

export function beginRequestDiagnostic(path: string, method: "GET" | "POST") {
  if (!diagnosticsEnabled) return;
  const route = path.split("?", 1)[0];
  let operation: Operation;
  if (method === "GET" && route === "/api/v1/stores") operation = "stores_get";
  else if (/^\/api\/v1\/stores\/[^/]+\/published-listings$/.test(route)) {
    operation = method === "GET" ? "listings_get" : "sync_post";
  } else if (method === "GET" && /^\/api\/v1\/stores\/[^/]+\/published-listings\/[^/]+\/description$/.test(route)) {
    operation = "description_get";
  } else if (method === "POST" && /^\/api\/v1\/stores\/[^/]+\/connection-check$/.test(route)) {
    operation = "store_token_check";
  } else if (/^\/api\/v1\/stores\/[^/]+\/orders(?:\/|$)/.test(route)) {
    if (method === "GET") {
      operation = route.endsWith("/sync-state") ? "order_state_get"
        : route.endsWith("/purchase") ? "order_purchase_get"
          : /^\/api\/v1\/stores\/[^/]+\/orders\/[^/]+$/.test(route) ? "order_detail_get" : "orders_get";
    } else if (route.endsWith("/sync")) operation = "orders_sync_post";
    else if (route.endsWith("/skip")) operation = "orders_skip_post";
    else if (route.endsWith("/reset")) operation = "orders_reset_post";
    else if (route.endsWith("/purchase")) operation = "order_purchase_post";
    else return;
  } else return;

  const session = lifetime;
  const requestId = ++nextRequestId;
  let start = performance.now();
  return (phase: Phase, response?: Response, code?: unknown, cacheHit = false) => {
    if (session !== lifetime) return;
    const end = performance.now();
    const timing = response?.headers.get("Server-Timing")?.split(",")
      .map((value) => value.trim())
      .filter((value) => /^(api|auth|read|sync);dur=\d{1,9}(\.\d{1,3})?$/.test(value))
      .slice(0, 4).join(", ");
    entries[cursor] = {
      time: new Date().toISOString(),
      request_id: requestId,
      operation,
      phase,
      duration_ms: Math.round((end - start) * 10) / 10,
      http_status: response?.status,
      error_code: typeof code === "string" && knownCodes.has(code) ? code :
        phase === "failed" ? failureCode(code, response) : undefined,
      token_cached: phase === "token_preflight" ? cacheHit : undefined,
      cache_hit: phase === "response_body" ? cacheHit : undefined,
      server_timing: timing || undefined,
    };
    start = end;
    cursor = (cursor + 1) % limit;
    changed();
  };
}

// Emitted by page lifecycle effects, never during React render.
export function recordCacheHit(path: string) {
  beginRequestDiagnostic(path, "GET")?.("response_body", undefined, undefined, true);
}

function failureCode(code: unknown, response?: Response) {
  switch (code) {
    case "aborted":
    case "timeout":
    case "token_preflight":
    case "invalid_body":
      return code;
    default:
      return response && !response.ok ? "http_" + response.status : "network";
  }
}

function subscribe(listener: () => void) {
  listeners.add(listener);
  return () => { listeners.delete(listener); };
}

export function diagnosticsJSON() {
  const ordered = entries.length === limit ? [...entries.slice(cursor), ...entries.slice(0, cursor)] : entries;
  return JSON.stringify(ordered, null, 2);
}

export default function DevelopmentDiagnostics() {
  useSyncExternalStore(subscribe, () => version);
  const [notice, setNotice] = useState("");
  const [exporting, setExporting] = useState(false);
  const json = diagnosticsJSON();

  async function copy() {
    try {
      await navigator.clipboard.writeText(json);
      setNotice("已复制诊断记录");
    } catch {
      setNotice("复制失败，可导出 JSON");
    }
  }

  async function download() {
    if (!entries.length || exporting) return;
    setExporting(true);
    setNotice("");
    try {
      // WebView2 supports the File System Access picker; unlike a synthetic
      // blob-link click, this reports when a real file has been written.
      type Handle = { createWritable: () => Promise<{
        write: (value: string) => Promise<void>; close: () => Promise<void>;
      }> };
      const host = window as Window & { showSaveFilePicker?: (options: {
        suggestedName: string; types: { description: string; accept: Record<string, string[]> }[];
      }) => Promise<Handle> };
      if (!host.showSaveFilePicker) {
        // WebKit-based webviews may lack the picker. Never report "saved" when
        // we could only copy the JSON; make the manual fallback explicit.
        await navigator.clipboard.writeText(json);
        setNotice("此客户端尚不支持系统文件保存，诊断 JSON 已复制。请粘贴到记事本，另存为 una-seller-diagnostics.json");
        return;
      }
      const file = await host.showSaveFilePicker({
        suggestedName: "una-seller-diagnostics.json",
        types: [{ description: "JSON 诊断日志", accept: { "application/json": [".json"] } }],
      });
      const writable = await file.createWritable();
      await writable.write(json);
      await writable.close();
      setNotice("已成功导出 una-seller-diagnostics.json");
    } catch (error) {
      setNotice(error instanceof Error && error.name === "AbortError"
        ? "已取消导出"
        : "导出未完成：" + (error instanceof Error ? error.message : "系统拒绝文件保存") +
          "。可点击「复制 JSON」手动保存。");
    } finally {
      setExporting(false);
    }
  }

  return <details>
    <summary>开发诊断（{entries.length}/{limit}）</summary>
    <div>
      <button type="button" onClick={copy} disabled={!entries.length}>复制 JSON</button>
      <button type="button" onClick={() => void download()} disabled={!entries.length || exporting}>{exporting ? "正在导出…" : "导出 JSON"}</button>
      <button type="button" onClick={() => { clearDiagnostics(); setNotice("已清空"); }} disabled={!entries.length}>清空</button>
    </div>
    <p role="status">{notice}</p>
    <pre tabIndex={0} aria-label="诊断记录" style={{ maxHeight: 320, overflow: "auto" }}>{json}</pre>
  </details>;
}
