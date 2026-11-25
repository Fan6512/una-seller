// Pure view-model helpers. No change to the saved listing_fields JSON contract.
type GalleryImage = { url: string; type: "main" | "additional" };

function withFirstAsMain(images: GalleryImage[]): GalleryImage[] {
  return images.map((image, index) => ({
    ...image, type: (index === 0 ? "main" : "additional") as "main" | "additional",
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

export function planSecondaryImageReuse(source: readonly string[], target: readonly string[]): { urls: string[]; error?: string; changed: boolean } {
  const valid = (url: string | undefined) => Boolean(url && /^https:\/\/[^\s]+$/i.test(url) && url.length <= 2000);
  const unchanged = (error: string) => ({ urls: [...target], error, changed: false });
  if (!valid(target[0])) return unchanged("目标变体尚未设置有效主图");
  if (!valid(source[0]) || source.length < 2 || source.slice(1).some(url => !valid(url))) return unchanged("来源变体没有有效副图");
  const seen = new Set([target[0], source[0]]);
  const urls = [target[0]];
  for (const url of source.slice(1)) {
    if (seen.has(url)) continue;
    seen.add(url);
    urls.push(url);
  }
  if (urls.length === 1) return unchanged("没有可复用的不同副图");
  if (urls.length > 30) return unchanged("超过变体 30 张图片上限");
  return { urls, changed: urls.length !== target.length || urls.some((url, i) => url !== target[i]) };
}
