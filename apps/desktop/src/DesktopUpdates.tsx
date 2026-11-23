import { useCallback, useEffect, useRef, useState, type RefObject } from "react";
import { invoke, isTauri } from "@tauri-apps/api/core";
import { check, type Update } from "@tauri-apps/plugin-updater";
import { formatDesktopVersion } from "./desktopVersion";

type Phase = "checking" | "idle" | "downloading" | "ready" | "installing";
type Activity = RefObject<{ syncs: number; installing: boolean }>;
const TEST_UPDATES_KEY = "una-seller:include-test-updates";
function readTestUpdatePreference(): boolean {
  try { return window.localStorage.getItem(TEST_UPDATES_KEY) === "1"; }
  catch { return false; }
}

export function useDesktopUpdates(activity: Activity) {
  const started = useRef(false);
  const running = useRef(false);
  // The updater may be disabled for unsigned builds; the installed package version isn't.
  const [installedVersion, setInstalledVersion] = useState<string | null>(null);
  const [installedVersionError, setInstalledVersionError] = useState(false);
  const [version, setVersion] = useState<string | null>(null);
  const [phase, setPhase] = useState<Phase>("checking");
  const [update, setUpdate] = useState<Update | null>(null);
  const [current, setCurrent] = useState(false);
  const [error, setError] = useState("");
  const [progress, setProgress] = useState({ received: 0, total: 0 });
  const [includeTestUpdates, setIncludeTestUpdates] = useState(readTestUpdatePreference);

  const checkForUpdate = useCallback(async (previous?: Update | null, testUpdates = includeTestUpdates) => {
    if (running.current) return;
    running.current = true;
    setPhase("checking");
    setError("");
    setCurrent(false);
    try {
      if (previous) {
        setUpdate(null);
        await previous.close();
      }
      const installed = isTauri() ? await invoke<string | null>("desktop_update_version") : null;
      setVersion(installed);
      if (!installed) return;
      const next = await check({
        timeout: 30_000,
        headers: { "X-Una-Include-Prerelease": testUpdates ? "1" : "0" },
      });
      setUpdate(next);
      setCurrent(!next);
    } catch (error) {
      console.error("Desktop update check failed", error);
      setError("检查更新失败。请检查网络后重试。");
    } finally {
      running.current = false;
      setPhase("idle");
    }
  }, [includeTestUpdates]);

  useEffect(() => {
    if (started.current) return;
    started.current = true;
    void checkForUpdate();
  }, [checkForUpdate]);

  useEffect(() => {
    if (!isTauri()) return;
    let mounted = true;
    void invoke<string>("desktop_app_version")
      .then(value => { if(mounted)setInstalledVersion(value); })
      .catch(() => { if(mounted)setInstalledVersionError(true); });
    return () => { mounted = false; };
  }, []);

  function changeTestUpdates(enabled: boolean) {
    try {
      window.localStorage.setItem(TEST_UPDATES_KEY, enabled ? "1" : "0");
    } catch { /* Preferences are optional; still apply for this session. */ }
    setIncludeTestUpdates(enabled);
    // Clear any previous channel's candidate before checking the new setting.
    if (phase === "idle" || phase === "ready") void checkForUpdate(update, enabled);
  }

  async function download() {
    if (!update || phase !== "idle" || running.current) return;
    running.current = true;
    setError("");
    setProgress({ received: 0, total: 0 });
    setPhase("downloading");
    try {
      await update.download((event) => {
        if (event.event === "Started") {
          setProgress({ received: 0, total: event.data.contentLength ?? 0 });
        } else if (event.event === "Progress") {
          setProgress(({ received, total }) => ({ received: received + event.data.chunkLength, total }));
        }
      }, { timeout: 180_000 });
      setPhase("ready");
    } catch (error) {
      console.error("Desktop update download failed", error);
      setError("下载或签名验证失败。请重新检查更新后重试。");
      setPhase("idle");
    } finally {
      running.current = false;
    }
  }

  async function install() {
    if (!update || phase !== "ready" || running.current || activity.current.syncs > 0) return;
    running.current = true;
    activity.current.installing = true;
    setError("");
    setPhase("installing");
    try {
      await update.install();
    } catch (error) {
      console.error("Desktop update install failed", error);
      setError("无法启动更新安装。请重新检查更新后重试。");
      setPhase("idle");
    } finally {
      activity.current.installing = false;
      running.current = false;
    }
  }

  return { installedVersion, installedVersionError, version, phase, update, current, error, progress, includeTestUpdates, changeTestUpdates, checkForUpdate, download, install };
}

export default function DesktopUpdates({ updates, syncCount, onLater }: {
  updates: ReturnType<typeof useDesktopUpdates>; syncCount: number; onLater: () => void;
}) {
  const { installedVersion, installedVersionError, version, phase, update, current, error, progress, includeTestUpdates, changeTestUpdates, checkForUpdate, download, install } = updates;
  const busy = phase === "checking" || phase === "downloading" || phase === "installing";

  return <>
    <div className="page-heading"><h1>设置</h1></div>
    <section className="detail-section" aria-label="客户端更新">
      <h2>客户端更新</h2>
      <p role="status">当前版本：{installedVersion ? formatDesktopVersion(installedVersion) : (installedVersionError?"无法读取版本":isTauri()?"版本读取中…":"开发预览")}</p>
      <label className="desktop-test-updates-toggle">
        <input type="checkbox" checked={includeTestUpdates}
          disabled={busy}
          onChange={e=>changeTestUpdates(e.target.checked)}/>
        <span>接收测试版更新</span>
      </label>
      <p className="listing-fields-note">默认关闭，只检查已发布的正式版本；开启后会同时检查正式版与 Pre-release，仅提示版本更高的更新。测试版同样必须通过签名验证。</p>
      <p className="page-actions">
        <button type="button" disabled={busy || phase === "ready" || (!version && !error)} onClick={() => void checkForUpdate(update)}>检查更新</button>
      </p>
      {phase === "checking" && <p role="status">正在检查更新…</p>}
      {!version && !busy && !error && <p>此版本未启用在线更新。</p>}
      {current && !busy && <p role="status">已是最新版本。</p>}
      {error && <p role="alert">{error}</p>}
      {update && <>
        <h3>可用版本 {formatDesktopVersion(update.version)}</h3>
        {update.body && <div className="description-text">{update.body}</div>}
        {phase === "downloading" && <p role="status">
          <progress aria-label="更新下载进度" max={progress.total || undefined}
            value={progress.total ? Math.min(progress.received, progress.total) : undefined} />{" "}
          {progress.total ? Math.min(100, Math.floor(progress.received / progress.total * 100)) + "%"
            : (progress.received / 1024 / 1024).toFixed(1) + " MB"}
        </p>}
        {phase === "installing" && <p role="status">正在启动安装…</p>}
        {phase === "ready" && syncCount > 0 && <p role="status">商品同步结束后可安装更新。</p>}
        {!busy && <p className="page-actions">
          {phase === "ready"
            ? <button className="primary-action" type="button" disabled={syncCount > 0} onClick={() => void install()}>安装并重启</button>
            : <button className="primary-action" type="button" onClick={() => void download()}>下载更新</button>}
          {" "}<button type="button" onClick={onLater}>稍后安装</button>
        </p>}
      </>}
    </section>
  </>;
}
