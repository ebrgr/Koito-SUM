#!/usr/bin/env python3
"""Verify fixture isolation, refusal to overwrite and fixed expected data."""

import sqlite3
from contextlib import closing
import tempfile
import unittest
from pathlib import Path

from roadmap06_fixture import check_database, create


class FixtureTest(unittest.TestCase):
    """Exercise the generated databases without opening an operational database."""

    def test_fixture_variants_and_isolated_mutation(self):
        """Editing one snapshot must not affect its siblings or the canonical base."""
        with tempfile.TemporaryDirectory(prefix="koito-022-test-") as directory:
            root = Path(directory) / "fixture"
            report = create(root)
            self.assertEqual(report["base_counts"], {"alpha": 7, "beta": 2, "empty": 0})
            self.assertEqual(report["large_counts"]["alpha"], 1009)
            with closing(sqlite3.connect(root / "mutation" / "koito.db")) as conn:
                conn.execute("DELETE FROM listens WHERE profile_id=1")
                conn.commit()
            self.assertEqual(check_database(root / "mutation" / "koito.db", {"alpha": 0, "beta": 2, "empty": 0})["alpha"], 0)
            for name in ("base.sqlite3", "go/koito.db", "react/koito.db"):
                check_database(root / name, {"alpha": 7, "beta": 2, "empty": 0})
            check_database(root / "zero" / "koito.db", {})
            check_database(root / "one" / "koito.db", {"alpha": 7})
            check_database(root / "large" / "koito.db", {"alpha": 1009, "beta": 2, "empty": 0})

    def test_existing_directory_is_never_modified(self):
        """Refuse an existing target before touching any of its files."""
        with tempfile.TemporaryDirectory(prefix="koito-022-test-") as directory:
            root = Path(directory) / "existing"
            root.mkdir()
            marker = root / "keep"
            marker.write_text("unchanged")
            with self.assertRaises(FileExistsError):
                create(root)
            self.assertEqual(marker.read_text(), "unchanged")


if __name__ == "__main__":
    unittest.main()
