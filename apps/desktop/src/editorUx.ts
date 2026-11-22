// Pure view-model helpers. No change to the saved listing_fields JSON contract.
type GalleryImage = { url: string; type: "main" | "additional" };

function withFirstAsMain(images: GalleryImage[]): GalleryImage[] {
  return images.map((image, index) => ({
    ...image, type: index === 0 ? "main" : "additional",
  }));
}

export function moveGalleryImage(images: GalleryImage[], from: number, to: number): GalleryImage[] {
  if (from < 0 || to < 0 || from >= images.length || to >= images.length || from === to) return images;
  const next = images.slice();
  const [chosen] = next.splice(from, 1);
  next.splice(to, 0, chosen);
  return withFirstAsMain(next);
}

// Only exact URL duplicates, ignoring surrounding whitespace. Different color
// photos and different URLs are intentionally never treated as duplicates.
export function removeDuplicateGalleryLinks(images: GalleryImage[]): GalleryImage[] {
  const seen = new Set<string>();
  return withFirstAsMain(images.filter(image => {
    const key = image.url.trim();
    if (!key) return true;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  }));
}

export function toggleVariantImage(urls: string[], url: string, checked: boolean): string[] {
  if (checked) return urls.includes(url) ? urls : [...urls, url];
  return urls.filter(item => item !== url);
}

export function renameVariantAttribute(
  attributes: Record<string, string>, oldKey: string, newKey: string,
): Record<string, string> {
  // A duplicate destination must not silently replace a different attribute.
  if (newKey !== oldKey && Object.prototype.hasOwnProperty.call(attributes, newKey)) return attributes;
  return Object.fromEntries(Object.entries(attributes).map(([key, value]) =>
    key === oldKey ? [newKey, value] : [key, value]));
}
