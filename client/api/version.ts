import semver from "semver";

// A numeric prerelease segment keeps RC10 newer than RC9.
export function normalizedVersion(value: string): string | null {
  return semver.valid(
    value
      .replace(/^(?:v)?(\d+\.\d+\.\d+)-rc\.([1-9]\d*)$/i, "$1-SUM-RC$2")
      .replace(/^v?(\d+\.\d+\.\d+)-SUM-RC([1-9]\d*)$/, "$1-SUM.RC.$2"),
  );
}
export function isNewerVersion(current: string, latest: string): boolean {
  const a = normalizedVersion(current);
  const b = normalizedVersion(latest);
  return !!a && !!b && semver.gt(b, a);
}
export interface ReleaseInfo {
  tag_name: string;
  name: string;
  body: string;
  html_url: string;
  draft: boolean;
}
export async function latestRelease(): Promise<ReleaseInfo | undefined> {
  // /latest excludes prereleases, which are used by SUM RC builds.
  const response = await fetch(
    "https://api.github.com/repos/ebrgr/Koito-SUM/releases?per_page=100",
  );
  if (!response.ok) return undefined;
  const releases: ReleaseInfo[] = await response.json();
  return releases
    .filter((r) => !r.draft && normalizedVersion(r.tag_name))
    .sort((a, b) =>
      semver.rcompare(
        normalizedVersion(a.tag_name)!,
        normalizedVersion(b.tag_name)!,
      ),
    )[0];
}
