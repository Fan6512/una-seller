// Layout depends on the actual gallery width, not the window or an assumed row size.
export const GALLERY_MIN_CARD = 170;
export const GALLERY_GAP = 12;

export function galleryColumns(width: number): number {
  return Math.max(1, Math.floor((Math.max(0, width) + GALLERY_GAP) / (GALLERY_MIN_CARD + GALLERY_GAP)));
}

// Insert after the entire current row, including a short final row.
export function expandedRowEnd(index: number, columns: number, count: number): number {
  if (index < 0 || index >= count) return -1;
  return Math.min(count - 1, (Math.floor(index / Math.max(1, columns)) + 1) * Math.max(1, columns) - 1);
}
