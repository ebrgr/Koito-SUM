#!/usr/bin/env python3
"""Check or increment the shared SUM RC version before committing."""
import argparse
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parent.parent
SOURCE = ROOT / "internal/version/VERSION"
FILES = ["client/package.json", "Dockerfile", "deploy/compose.yaml", "deploy/.env.example",
         "README.md", "DEPLOYMENT.md", "MULTIUSER_DESIGN.md", "MULTIUSER_DESIGN.en.md"]
PATTERN = r"(\d+\.\d+\.\d+)-SUM-RC([1-9]\d*)"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bump", action="store_true", help="Increment RC and synchronize current documentation/deployment defaults")
    parser.add_argument("--check-parent", action="store_true", help="Require a higher RC than the parent commit when it has a VERSION file")
    args = parser.parse_args()
    current = SOURCE.read_text().strip()
    match = re.fullmatch(PATTERN, current)
    if not match:
        raise SystemExit("Invalid version: " + current)
    contents = {name: (ROOT / name).read_text() for name in FILES}
    for name, text in contents.items():
        if current not in text:
            raise SystemExit(f"Version out of sync: {name}")
    if json.loads(contents["client/package.json"])["version"] != current:
        raise SystemExit("Frontend package version out of sync")
    if args.check_parent:
        parent = subprocess.run(["git", "show", "HEAD^:internal/version/VERSION"], cwd=ROOT, capture_output=True, text=True)
        if parent.returncode == 0:
            previous = re.fullmatch(PATTERN, parent.stdout.strip())
            def key(m):
                return (*map(int, m[1].split(".")), int(m[2]))
            if previous and match[1] != previous[1]:
                raise SystemExit("Preserve the upstream base version; increment only the SUM RC number")
            if previous and key(match) <= key(previous):
                raise SystemExit("Increment the version before committing: make version.bump")
    if args.bump:
        next_version = f"{match[1]}-SUM-RC{int(match[2]) + 1}"
        for name, text in contents.items():
            (ROOT / name).write_text(text.replace(current, next_version))
        notes = ROOT / f"releases/{current}.md"
        if notes.exists():
            (ROOT / f"releases/{next_version}.md").write_text(notes.read_text().replace(current, next_version))
        SOURCE.write_text(next_version + "\n")
        print(f"Koito {next_version} (previous release notes preserved; review the new notes before committing)")
    else:
        print(f"Koito {current}: version files synchronized")


if __name__ == "__main__":
    main()
