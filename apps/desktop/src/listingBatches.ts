export type ListingInterruption = { code: string; platform_status?: number; retryable: boolean };
export type ListingBatch = {
  saved: number; processed: number; skipped: number; needs_review: number; partial: number;
  next_offset: number | null; scan_required: boolean; interruption?: ListingInterruption;
};

export async function runListingBatches(
  offset: number,
  restart: boolean,
  signal: AbortSignal,
  request: (offset: number, restart: boolean, signal: AbortSignal) => Promise<ListingBatch>,
  progress: (batch: ListingBatch) => void,
) {
  for (let count = 0; count < 10; count++) {
    signal.throwIfAborted();
    const batch = await request(offset, restart && count === 0, signal);
    signal.throwIfAborted();
    progress(batch);
    if (batch.interruption || batch.next_offset == null) return;
    if (batch.next_offset <= offset) throw new Error("同步位置未前进，请刷新后继续");
    offset = batch.next_offset;
  }
}
