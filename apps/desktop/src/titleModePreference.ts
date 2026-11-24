import { useSyncExternalStore } from "react";

// UI-only preference. Never used as Mercado Libre publishing authorization;
// no secret or product data is stored with this preference.
export type SiteTitleMode = "platform_auto" | "localized";
const storageKey = "una-seller:site-title-mode:v1";
const changeEvent = "una-seller:site-title-mode-changed";
let fallback: SiteTitleMode = "localized";

export function getSiteTitleMode(): SiteTitleMode {
  try {
    const value = window.localStorage.getItem(storageKey);
    return value === "platform_auto" || value === "localized" ? value : fallback;
  } catch {
    return fallback;
  }
}

export function setSiteTitleMode(mode: SiteTitleMode): void {
  fallback = mode;
  try {
    window.localStorage.setItem(storageKey, mode);
  } catch {
    // Client may block localStorage; keep this session usable in memory.
  }
  window.dispatchEvent(new Event(changeEvent));
}

function subscribeSiteTitleMode(listener: () => void): () => void {
  window.addEventListener(changeEvent, listener);
  window.addEventListener("storage", listener);
  return () => {
    window.removeEventListener(changeEvent, listener);
    window.removeEventListener("storage", listener);
  };
}

export function useSiteTitleMode(): SiteTitleMode {
  return useSyncExternalStore(subscribeSiteTitleMode, getSiteTitleMode, () => "localized");
}
