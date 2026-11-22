// Pure workbench decisions: keep approval status and retry behavior testable
// independently of React, OAuth, and the Tauri runtime.
export type ApprovalOutcome = "created" | "already_approved" | "failed";

export function countApprovalOutcomes(outcomes: readonly ApprovalOutcome[]) {
  return {
    created: outcomes.filter(x => x === "created").length,
    alreadyApproved: outcomes.filter(x => x === "already_approved").length,
    failed: outcomes.filter(x => x === "failed").length,
  };
}

export function productSaveNeedsReadback(status: number | undefined): boolean {
  // Validation/revision/not-found errors are known rejections, NOT unknown writes.
  return status === undefined || ![400,404,409,415].includes(status);
}

export function emptyProductPageFallback(offset: number, itemCount: number, pageSize: number): number | null {
  return offset > 0 && itemCount === 0 ? Math.max(0, offset - pageSize) : null;
}
