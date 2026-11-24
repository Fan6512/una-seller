import { clearDiagnostics, diagnosticsEnabled } from "./DevelopmentDiagnostics";

// Lightweight client for the open-source Supabase Auth (GoTrue) REST API.
// Only the publishable key ships with the desktop; no DB or server secret is embedded.
const authBase = (import.meta.env.VITE_SUPABASE_URL || "https://rwktwfxzfkzglxwaulcm.supabase.co").replace(/\/$/, "");
const publishableKey = import.meta.env.VITE_SUPABASE_PUBLISHABLE_KEY || "sb_publishable_KUDh1isf3-EknWx-rviV0A_w3lm-u0z";
const storageKey = "una-seller.auth.v1";

export type AuthSession = {
  access_token: string;
  refresh_token: string;
  expires_at: number; // UNIX seconds
  user: { id: string; email: string };
};

type TokenResponse = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: { id: string; email?: string };
};

let current: AuthSession | null | undefined;
let refreshInFlight: Promise<AuthSession> | null = null;
let sessionLifetime = new AbortController();

export function authLifetime(): AbortSignal {
  return sessionLifetime.signal;
}

function endSessionLifetime() {
  sessionLifetime.abort();
  sessionLifetime = new AbortController();
  refreshInFlight = null;
}

function isSession(value: unknown): value is AuthSession {
  if (!value || typeof value !== "object") return false;
  const v = value as Partial<AuthSession>;
  return typeof v.access_token === "string" && Boolean(v.access_token)
    && typeof v.refresh_token === "string" && Boolean(v.refresh_token)
    && typeof v.expires_at === "number" && Number.isFinite(v.expires_at)
    && typeof v.user?.id === "string" && typeof v.user.email === "string";
}

function loadSession(): AuthSession | null {
  if (current !== undefined) return current;
  try {
    const raw = localStorage.getItem(storageKey);
    const value: unknown = raw ? JSON.parse(raw) : null;
    current = isSession(value) ? value : null;
  } catch {
    current = null;
  }
  return current;
}

function storeSession(value: AuthSession) {
  current = value;
  localStorage.setItem(storageKey, JSON.stringify(value));
}

function clearSession(notify = false) {
  if (diagnosticsEnabled) clearDiagnostics();
  endSessionLifetime();
  current = null;
  localStorage.removeItem(storageKey);
  if (notify) window.dispatchEvent(new Event("una-seller:auth-lost"));
}

async function tokenRequest(grantType: "password" | "refresh_token", payload: object): Promise<AuthSession> {
  let response: Response;
  try {
    response = await fetch(authBase + "/auth/v1/token?grant_type=" + grantType, {
      method: "POST",
      signal: AbortSignal.timeout(8_000),
      headers: { "apikey": publishableKey, "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      cache: "no-store",
    });
  } catch {
    throw new Error("无法连接登录服务，请检查网络后重试");
  }
  if (!response.ok) {
    if (response.status === 429) throw new Error("登录请求过于频繁，请稍后重试");
    if (grantType === "password") throw new Error("邮箱或密码错误，或账号尚未完成验证");
    throw new Error("登录已过期，请重新登录");
  }
  const data = (await response.json()) as TokenResponse;
  if (!data.access_token || !data.refresh_token || !data.user?.id || !Number.isFinite(data.expires_in)) {
    throw new Error("登录服务返回了无效的会话");
  }
  return {
    access_token: data.access_token,
    refresh_token: data.refresh_token,
    expires_at: Math.floor(Date.now() / 1000) + data.expires_in,
    user: { id: data.user.id, email: data.user.email || "" },
  };
}

export async function signIn(email: string, password: string): Promise<AuthSession> {
  const next = await tokenRequest("password", { email: email.trim(), password });
  endSessionLifetime();
  if (diagnosticsEnabled) clearDiagnostics();
  storeSession(next);
  return next;
}

export async function getAccessToken(forceRefresh = false, onPreflight?: (cached: boolean) => void): Promise<string> {
  const session = loadSession();
  if (!session) throw new Error("请先登录");
  if (!forceRefresh && session.expires_at > Math.floor(Date.now() / 1000) + 60) {
    onPreflight?.(true);
    return session.access_token;
  }
  const lifetime = authLifetime();
  if (!refreshInFlight) {
    refreshInFlight = tokenRequest("refresh_token", { refresh_token: session.refresh_token })
      .then((next) => {
        // Logout may have occurred while refreshing.
        if (next.user.id !== session.user.id) throw new Error("登录已过期，请重新登录");
        if (!lifetime.aborted && loadSession()?.refresh_token === session.refresh_token) storeSession(next);
        return next;
      }).finally(() => { if (!lifetime.aborted) refreshInFlight = null; });
  }
  try {
    await refreshInFlight;
    lifetime.throwIfAborted();
    onPreflight?.(false);
    return loadSession()!.access_token;
  } catch (error) {
    // Network failures are recoverable; only an explicit expired-session error logs out.
    if (!lifetime.aborted && error instanceof Error && error.message === "登录已过期，请重新登录") clearSession(true);
    throw error;
  }
}

export async function restoreSession(): Promise<AuthSession | null> {
  const session = loadSession();
  if (!session) return null;
  try {
    await getAccessToken();
    return loadSession();
  } catch {
    // A transient network issue must not destroy a valid persisted refresh token.
    return loadSession();
  }
}

export async function signOut(): Promise<void> {
  const token = loadSession()?.access_token;
  clearSession();
  if (!token) return;
  try {
    await fetch(authBase + "/auth/v1/logout", {
      method: "POST",
      signal: AbortSignal.timeout(8_000),
      headers: { "apikey": publishableKey, "Authorization": "Bearer " + token },
    });
  } catch {
    // Local logout still clears credentials if the network is offline.
  }
}

// Supabase Auth owns password hashes and optional email OTP reauthentication.
// The desktop never saves a current/new password or OTP to persistent storage.
export async function sendPasswordVerificationNonce(): Promise<void> {
  const token = await getAccessToken();
  const response = await fetch(authBase + "/auth/v1/reauthenticate", {
    method: "GET",
    cache: "no-store",
    signal: AbortSignal.timeout(10_000),
    headers: { "apikey": publishableKey, "Authorization": "Bearer " + token },
  });
  if (!response.ok) throw new Error("无法发送密码修改验证码，请检查邮箱与登录状态");
}

export async function changeAccountPassword(currentPassword: string, newPassword: string, nonce = ""): Promise<void> {
  const token = await getAccessToken();
  const response = await fetch(authBase + "/auth/v1/user", {
    method: "PUT",
    cache: "no-store",
    signal: AbortSignal.timeout(10_000),
    headers: {
      "apikey": publishableKey,
      "Authorization": "Bearer " + token,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      current_password: currentPassword,
      password: newPassword,
      ...(nonce.trim() ? { nonce: nonce.trim() } : {}),
    }),
  });
  if (response.ok) return;
  const data = await response.json().catch(() => null) as { code?: string; error_code?: string } | null;
  const errorCode = data?.code ?? data?.error_code;
  if (errorCode === "reauthentication_needed") throw new Error("当前账号要求邮箱二次验证，请先发送验证码后填写。");
  if (errorCode === "same_password") throw new Error("新密码不能与当前密码相同。");
  if (response.status === 401) throw new Error("当前登录已过期，请重新登录后修改密码。");
  if (response.status === 422 || response.status === 400) throw new Error("当前密码、验证码或新密码不符合登录服务要求，请核对后再试。");
  throw new Error("修改密码失败，请稍后重试。");
}

export function rejectSession() {
  clearSession(true);
}
