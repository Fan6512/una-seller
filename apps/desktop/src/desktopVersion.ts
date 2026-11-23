// The installed Tauri package uses signed SemVer internally.
// Present only MAJOR.MINOR to sellers; a "-beta.N" prerelease is still a test.
export function formatDesktopVersion(version: string): string {
  const match=/^(\d+)\.(\d+)\.(\d+)(-beta\.[1-9]\d*)?$/.exec(version);
  if (!match) return "V"+version;
  const [,major,minor,,beta]=match;
  return "V"+major+"."+minor+(major==="0"||beta?" 测试版":" 正式版");
}
