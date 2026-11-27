import { useEffect, useLayoutEffect, useRef } from "react";

/**
 * Display the exact new value immediately. A brief opacity cue explains a
 * discrete change without fake intermediate quantities or changing layout.
 * Use for low-frequency aggregate counters only, not typing or table rows.
 */
export function MotionValue({ value }: { value: string | number }) {
  const ref = useRef<HTMLSpanElement>(null);
  const previous = useRef(value);
  const running = useRef<Animation | null>(null);

  useLayoutEffect(() => {
    if (Object.is(previous.current, value)) return;
    previous.current = value;
    const node = ref.current;
    if (!node || typeof node.animate !== "function" ||
        window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      running.current?.cancel();
      running.current = null;
      return;
    }
    // Read only this small label on semantic change, not the full document.
    const style = getComputedStyle(node);
    const from = running.current?.playState === "running" ? Number(style.opacity) : .76;
    const duration = Number.parseFloat(style.getPropertyValue("--motion-standard")) || 160;
    const easing = style.getPropertyValue("--motion-ease").trim() || "ease-out";
    running.current?.cancel();
    running.current = node.animate([{ opacity: from }, { opacity: 1 }], { duration, easing });
  }, [value]);

  useEffect(() => {
    const media = window.matchMedia("(prefers-reduced-motion: reduce)");
    const stop = () => {
      if (media.matches) {
        running.current?.cancel();
        running.current = null;
      }
    };
    media.addEventListener("change", stop);
    return () => {
      media.removeEventListener("change", stop);
      running.current?.cancel();
      running.current = null;
    };
  }, []);

  return <span ref={ref} className="motion-value">{value}</span>;
}
