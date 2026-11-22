// The installed Tauri package uses SemVer MAJOR.MINOR.PATCH internally.
// Show only MAJOR.MINOR in the UI, even when an older package used a patch.
export function formatDesktopVersion(version: string): string {
  const match=/^(\d+)\.(\d+)\.(\d+)$/.exec(version);
  if(!match)return "V"+version;
  const [,major,minor]=match;
  return "V"+major+"."+minor+(major==="0"?" 测试版":" 正式版");
}
