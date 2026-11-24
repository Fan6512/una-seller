import { useSyncExternalStore } from "react";

// UI-only, device-local preference. Never publishing authorization.
export type SiteTitleMode = "platform_auto" | "localized";
const storageKey = "una-seller:site-title-mode:v1";
const changeEvent = "una-seller:site-title-mode-changed";
let fallback: SiteTitleMode = "localized";
let memoryOnly = false;

export function getSiteTitleMode(): SiteTitleMode {
  if (memoryOnly) return fallback;
  try {
    const value = window.localStorage.getItem(storageKey);
    fallback = value === "platform_auto" || value === "localized" ? value : "localized";
  } catch {
    // Keep the last usable preference when storage is unavailable.
  }
  return fallback;
}

export function setSiteTitleMode(mode: SiteTitleMode): void {
  if (mode !== "platform_auto" && mode !== "localized") return;
  fallback = mode;
  try {
    window.localStorage.setItem(storageKey, mode);
    memoryOnly = false;
  } catch {
    // A failed write must not let an older readable value undo this selection.
    memoryOnly = true;
  }
  window.dispatchEvent(new Event(changeEvent));
}

export function subscribeSiteTitleMode(listener: () => void): () => void {
  const onStorage = (event: StorageEvent) => {
    if (event.key !== storageKey && event.key !== null) return;
    if (event.storageArea && event.storageArea !== window.localStorage) return;
    memoryOnly = false;
    listener();
  };
  window.addEventListener(changeEvent, listener);
  window.addEventListener("storage", onStorage);
  return () => {
    window.removeEventListener(changeEvent, listener);
    window.removeEventListener("storage", onStorage);
  };
}

export function useSiteTitleMode(): SiteTitleMode {
  return useSyncExternalStore(subscribeSiteTitleMode, getSiteTitleMode, () => "localized");
}
