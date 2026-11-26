import { useLayoutEffect } from "react";
import { Toaster, toast } from "@yusufceng1z/ding-react";
import type { FeedbackView } from "./operationFeedback";

const capsuleId = "una-operation-feedback";
let layer: HTMLDivElement | null = null;
let lastModal: Element | null = null;
let queued: FeedbackView | null = null;
let frame = 0;

function present(view: FeedbackView) {
  if (!layer?.isConnected) return;
  // Only change top-layer ordering when a new native modal actually opens.
  if (layer.showPopover) {
    const modal = document.querySelector("dialog:modal");
    if (layer.matches(":popover-open")) {
      if (modal && modal !== lastModal) {
        layer.hidePopover();
        layer.showPopover();
      }
    } else {
      layer.showPopover();
    }
    lastModal = modal;
  }
  toast.show(view.kind, view.title, {
    id: capsuleId,
    description: view.description,
    meta: view.count && view.count > 1 ? "×" + view.count : undefined,
    progress: view.progress,
    duration: view.kind === "loading" ? 0 : 5000,
  });
}

export function renderOperationFeedback(view: FeedbackView | null) {
  if (!view) {
    if (frame) cancelAnimationFrame(frame);
    frame = 0;
    queued = null;
    lastModal = null;
    // Ding's singleton retains the last text in each live region after dismissal.
    const { status, alert } = toast.getState();
    toast.configure({ sound: false });
    if (status.trim()) toast.info("", { id: capsuleId, duration: 0 });
    if (alert.trim()) toast.error("", { id: capsuleId, duration: 0 });
    toast.dismiss(undefined, { immediate: true });
    return;
  }
  // Synchronous bursts (e.g. local variant reuse) need only their latest
  // visual state, not one spring restart and React render per item.
  queued = view;
  if (frame) return;
  frame = requestAnimationFrame(() => {
    frame = 0;
    const next = queued;
    queued = null;
    if (!next) return;
    try {
      present(next);
    } catch {
      // Deferred presentation must preserve the same failure isolation as
      // the business feedback boundary. Do not log customer content.
    }
  });
}

export function OperationFeedbackToaster({ sound = false }: { sound?: boolean }) {
  useLayoutEffect(() => {
    toast.configure({
      held: "replace", sound,
      springs: { enter: { stiffness: 360, damping: 36 }, morph: { stiffness: 360, damping: 36 } },
    });
  }, [sound]);
  // Mounted inside desktop-content: the popover inherits the shell sidebar width.
  // CSS constrains the viewport to that pane even in the native top layer.
  return <div ref={(element) => { layer = element; }} className="una-feedback-layer" popover="manual"
    style={{ inset: 0, margin: 0, border: 0, padding: 0, background: "transparent",
      overflow: "visible", pointerEvents: "none" }}>
    <Toaster position="top-center" theme="light"
      labels={{ region: "操作反馈", dismiss: "关闭操作反馈" }} />
  </div>;
}
