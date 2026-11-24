// Pure workbench decisions: keep approval status and retry behavior testable
// independently of React, OAuth, and the Tauri runtime.
export type ApprovalOutcome = "created" | "already_approved" | "failed";

export function approvalBlockReason(selectedCount: number, draftDirty: boolean): string | null {
  if (selectedCount === 0) return "请先勾选要转入产品箱的商品；可以勾选商品卡片，或点击「勾选当前商品」。";
  if (draftDirty) return "当前整理工作稿有未保存的修改，请先保存工作稿再转入产品箱。";
  return null;
}

export function countApprovalOutcomes(outcomes: readonly ApprovalOutcome[]) {
  return {
    created: outcomes.filter(x => x === "created").length,
    alreadyApproved: outcomes.filter(x => x === "already_approved").length,
    failed: outcomes.filter(x => x === "failed").length,
  };
}

export function productSaveNeedsReadback(status: number | undefined): boolean {
  // Validation/revision/not-found errors are known rejections, NOT unknown writes.
  return status === undefined || ![400,404,409,415,422].includes(status);
}

export function emptyProductPageFallback(offset: number, itemCount: number, pageSize: number): number | null {
  return offset > 0 && itemCount === 0 ? Math.max(0, offset - pageSize) : null;
}
