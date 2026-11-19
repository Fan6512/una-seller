import { useCallback, useEffect, useRef, useState, type RefObject } from "react";
import { invoke, isTauri } from "@tauri-apps/api/core";
import { check, type Update } from "@tauri-apps/plugin-updater";

type Phase = "checking" | "idle" | "downloading" | "ready" | "installing";
type Activity = RefObject<{ syncs: number; installing: boolean }>;

export function useDesktopUpdates(activity: Activity) {
  const started = useRef(false);
  const running = useRef(false);
  const [version, setVersion] = useState<string | null>(null);
  const [phase, setPhase] = useState<Phase>("checking");
  const [update, setUpdate] = useState<Update | null>(null);
  const [current, setCurrent] = useState(false);
  const [error, setError] = useState("");
  const [progress, setProgress] = useState({ received: 0, total: 0 });

  const checkForUpdate = useCallback(async (previous?: Update | null) => {
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
      const next = await check({ timeout: 15_000 });
      setUpdate(next);
      setCurrent(!next);
    } catch (error) {
      console.error("Desktop update check failed", error);
      setError("检查更新失败。请检查网络后重试。");
    } finally {
      running.current = false;
      setPhase("idle");
    }
  }, []);

  useEffect(() => {
    if (started.current) return;
    started.current = true;
    void checkForUpdate();
  }, [checkForUpdate]);

  async function download() {
    if (!update || running.current) return;
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

  return { version, phase, update, current, error, progress, checkForUpdate, download, install };
}

export default function DesktopUpdates({ updates, syncCount, onLater }: {
  updates: ReturnType<typeof useDesktopUpdates>; syncCount: number; onLater: () => void;
}) {
  const { version, phase, update, current, error, progress, checkForUpdate, download, install } = updates;
  const busy = phase === "checking" || phase === "downloading" || phase === "installing";

  return <>
    <div className="page-heading"><h1>设置</h1></div>
    <section className="detail-section" aria-label="客户端更新">
      <h2>客户端更新</h2>
      {version && <p>当前版本 {version}</p>}
      <p className="page-actions">
        <button type="button" disabled={busy || (!version && !error)} onClick={() => void checkForUpdate(update)}>检查更新</button>
      </p>
      {phase === "checking" && <p role="status">正在检查更新…</p>}
      {!version && !busy && !error && <p>此版本未启用在线更新。</p>}
      {current && !busy && <p role="status">已是最新版本。</p>}
      {error && <p role="alert">{error}</p>}
      {update && <>
        <h3>可用版本 {update.version}</h3>
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
