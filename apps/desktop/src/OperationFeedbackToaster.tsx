import { useLayoutEffect } from "react";
import { Toaster, toast } from "@yusufceng1z/ding-react";
import type { FeedbackView } from "./operationFeedback";

const capsuleId = "una-operation-feedback";
let layer: HTMLDivElement | null = null;

export function renderOperationFeedback(view: FeedbackView | null) {
  if (!view) {
    // Ding's singleton retains the last text in each live region after dismissal.
    const { status, alert } = toast.getState();
    toast.configure({ sound: false });
    if (status.trim()) toast.info("", { id: capsuleId, duration: 0 });
    if (alert.trim()) toast.error("", { id: capsuleId, duration: 0 });
    toast.dismiss(undefined, { immediate: true });
    return;
  }
  // Order details use a native modal dialog. Raise the capsule above its backdrop.
  if (layer?.showPopover) {
    const modal = document.querySelector("dialog:modal");
    if (layer.matches(":popover-open") && modal) {
      layer.hidePopover();
    }
    layer.showPopover();
  }
  toast.show(view.kind, view.title, {
    id: capsuleId,
    description: view.description,
    meta: view.count && view.count > 1 ? "×" + view.count : undefined,
    progress: view.progress,
    duration: view.kind === "loading" ? 0 : 5000,
  });
}

export function OperationFeedbackToaster({ sound = false }: { sound?: boolean }) {
  useLayoutEffect(() => { toast.configure({ held: "replace", sound }); }, [sound]);
  return <div ref={(element) => { layer = element; }} popover="manual"
    style={{ inset: 0, margin: 0, border: 0, padding: 0, background: "transparent",
      overflow: "visible", pointerEvents: "none" }}>
    <Toaster position="top-center" theme="light"
      labels={{ region: "操作反馈", dismiss: "关闭操作反馈" }} />
  </div>;
}
