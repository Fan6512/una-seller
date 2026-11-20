import { useState, useSyncExternalStore } from "react";

export const diagnosticsEnabled = import.meta.env.VITE_UNA_SELLER_DEV_DIAGNOSTICS === "1";

type Operation = "stores_get" | "listings_get" | "description_get" | "store_token_check" | "sync_post";
type Phase = "token_preflight" | "response_headers" | "response_body" | "failed";
type Entry = {
  time: string;
  operation: Operation;
  phase: Phase;
  duration_ms: number;
  http_status?: number;
  error_code?: string;
  cache_hit: boolean;
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
  "listing_identity_mismatch", "listing_review_changed", "listing_canceled", "listing_invalid_response",
]);
let cursor = 0;
let version = 0;
let lifetime = 0;

function changed() {
  version++;
  for (const listener of listeners) listener();
}

export function clearDiagnostics() {
  lifetime++;
  cursor = 0;
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
  } else return;

  const session = lifetime;
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
      operation,
      phase,
      duration_ms: Math.round((end - start) * 10) / 10,
      http_status: response?.status,
      error_code: typeof code === "string" && knownCodes.has(code) ? code :
        phase === "failed" ? failureCode(code, response) : undefined,
      cache_hit: cacheHit,
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
  const json = diagnosticsJSON();

  async function copy() {
    try {
      await navigator.clipboard.writeText(json);
      setNotice("已复制诊断记录");
    } catch {
      setNotice("复制失败，可导出 JSON");
    }
  }

  function download() {
    const url = URL.createObjectURL(new Blob([json], { type: "application/json" }));
    const link = document.createElement("a");
    link.href = url;
    link.download = "una-seller-diagnostics.json";
    document.body.append(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 1_000);
  }

  return <details>
    <summary>开发诊断（{entries.length}/{limit}）</summary>
    <div>
      <button type="button" onClick={copy} disabled={!entries.length}>复制 JSON</button>
      <button type="button" onClick={download} disabled={!entries.length}>导出 JSON</button>
      <button type="button" onClick={() => { clearDiagnostics(); setNotice("已清空"); }} disabled={!entries.length}>清空</button>
    </div>
    <p role="status">{notice}</p>
    <pre tabIndex={0} aria-label="诊断记录" style={{ maxHeight: 320, overflow: "auto" }}>{json}</pre>
  </details>;
}
