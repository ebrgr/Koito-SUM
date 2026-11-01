import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { compileTypeScript } from "./typescript.mjs";
import { pathToFileURL } from "node:url";

const require = createRequire(import.meta.url);
const source = (
  await readFile(new URL("../api/version.ts", import.meta.url), "utf8")
).replace(
  '"semver"',
  JSON.stringify(pathToFileURL(require.resolve("semver")).href),
);
const { isNewerVersion, latestRelease } = await import(
  `data:text/javascript,${encodeURIComponent(compileTypeScript(source))}`
);

test("SUM release candidates compare numerically, including the previous tag format", () => {
  assert.equal(isNewerVersion("0.1.0-SUM-RC9", "0.1.0-SUM-RC10"), true);
  assert.equal(isNewerVersion("0.1.0-SUM-RC10", "0.1.0-SUM-RC9"), false);
  assert.equal(isNewerVersion("v0.1.0-rc.1", "0.1.0-SUM-RC2"), true);
  assert.equal(isNewerVersion("0.1.0-SUM-RC2", "v0.1.0-SUM-RC2"), false);
  assert.equal(isNewerVersion("0.1.0-SUM-RC2", "0.1.0"), true);
  assert.equal(isNewerVersion("0.1.0-SUM-RC2", "invalid"), false);
});

test("update lookup selects the highest published version and includes prereleases", async (t) => {
  t.mock.method(
    globalThis,
    "fetch",
    async () =>
      new Response(
        JSON.stringify([
          { tag_name: "0.1.0-SUM-RC9", draft: false, prerelease: true },
          { tag_name: "0.1.0-SUM-RC11", draft: true },
          { tag_name: "0.1.0-SUM-RC10", draft: false, prerelease: true },
          { tag_name: "dev", draft: false },
        ]),
      ),
  );
  assert.equal((await latestRelease()).tag_name, "0.1.0-SUM-RC10");
});
