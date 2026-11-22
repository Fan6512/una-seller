// Versions are read from the installed Tauri package, not inferred from CI runs.
// 0.x is a QA candidate; major >= 1 is an explicitly approved stable release.
export function formatDesktopVersion(version: string): string {
  const match=/^(\d+)\.(\d+)\.(\d+)$/.exec(version);
  if(!match)return "V"+version;
  const [,major,minor,patch]=match;
  const short=patch==="0"?"V"+major+"."+minor:"V"+major+"."+minor+"."+patch;
  return short+(major==="0"?" 测试版":" 正式版");
}
