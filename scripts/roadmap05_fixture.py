#!/usr/bin/env python3
"""Build an isolated SQLite v3 validation fixture with independently fixed totals."""

import contextlib
import json
import sqlite3
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
FIXED_NOW = "2025-06-01T12:00:00Z"
PROCESS_TZ = "UTC"
BASE = 1714521600  # 2024-05-01T00:00:00Z
IMAGE = "11111111-1111-4111-8111-111111111111"
EXPECTED = {"alpha": 3, "beta": 2, "empty": 0}


def build_fixture(conn, large=False):
    """Seed a new disposable database; callers must never supply an existing database."""
    if conn.execute("SELECT count(*) FROM sqlite_master").fetchone()[0]:
        raise ValueError("fixture requires an empty database")
    conn.execute("PRAGMA foreign_keys=ON")
    for name in ("001_initial_schema.sql", "002_unique_mbid.sql", "003_listening_profiles.sql"):
        sql = (ROOT / "db/migrations_sqlite" / name).read_text().split("-- +goose Down")[0]
        conn.executescript(sql)
    conn.execute("PRAGMA foreign_keys=ON")
    conn.executescript("""
        CREATE TABLE goose_db_version (id INTEGER PRIMARY KEY AUTOINCREMENT,
          version_id INTEGER NOT NULL, is_applied BOOLEAN NOT NULL,
          tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP);
        INSERT INTO goose_db_version(version_id,is_applied) VALUES (0,1),(1,1),(2,1),(3,1);
        INSERT INTO listening_profiles(id,slug,display_name,created_at) VALUES
          (1,'alpha','Alpha',1714521600),(2,'beta','Beta',1714521600),(3,'empty','Empty',1714521600);
        INSERT INTO artists(id,image) VALUES(1,'11111111-1111-4111-8111-111111111111'),
          (2,'11111111-1111-4111-8111-111111111111'),(3,NULL);
        INSERT INTO artist_aliases VALUES(1,'Artista comum','fixture',1),
          (1,'共通','fixture',0),(2,'Artista Beta','fixture',1),(3,'Sem escuta','fixture',1);
        INSERT INTO releases(id,image) VALUES(1,'11111111-1111-4111-8111-111111111111'),(2,NULL),(3,NULL);
        INSERT INTO release_aliases VALUES(1,'Álbum comum','fixture',1),
          (2,'Álbum Beta','fixture',1),(3,'Álbum vazio','fixture',1);
        INSERT INTO artist_releases VALUES(1,1,1),(2,1,0),(2,2,1),(3,3,1);
        INSERT INTO tracks(id,release_id,duration) VALUES(1,1,180),(2,2,0),(3,3,120),(4,1,200);
        INSERT INTO track_aliases VALUES(1,'Faixa comum','fixture',1),(1,'共通曲','fixture',0),
          (2,'Só Beta','fixture',1),(3,'Nunca tocada','fixture',1),(4,'Outra Alpha','fixture',1);
        INSERT INTO artist_tracks VALUES(1,1,1),(2,1,0),(2,2,1),(3,3,1),(1,4,1);
        INSERT INTO api_keys(id,key,profile_id,created_at,label) VALUES
          (1,'fixture-alpha-key',1,1714521600,'fixture'),(2,'fixture-beta-key',2,1714521600,'fixture');
    """)
    listens = [(1, 1, BASE), (1, 1, BASE + 1), (1, 4, BASE + 2),
               (2, 1, BASE), (2, 2, BASE + 3)]
    conn.executemany("INSERT INTO listens(profile_id,track_id,listened_at,client) VALUES(?,?,?,'fixture')", listens)
    if large:
        # Two tracks share each second, including the 1000-row export boundary.
        conn.executemany("INSERT INTO listens(profile_id,track_id,listened_at,client) VALUES(1,?,?, 'fixture-large')",
                         ((1 if i % 2 == 0 else 4, BASE + 100 + i // 2) for i in range(1002)))
    conn.commit()


@contextlib.contextmanager
def disposable_fixture(large=False):
    """Yield a fresh database path and delete its directory when the caller finishes."""
    with tempfile.TemporaryDirectory(prefix="koito-roadmap05-") as directory:
        path = Path(directory) / "koito.db"
        with sqlite3.connect(path) as conn:
            build_fixture(conn, large)
        yield path


def verify(path, large=False):
    """Check fixture facts using literal expectations and SQLite integrity checks."""
    expected = dict(EXPECTED)
    if large:
        expected["alpha"] += 1002
    with sqlite3.connect(path) as conn:
        conn.execute("PRAGMA foreign_keys=ON")
        actual = dict(conn.execute("SELECT p.slug,count(l.track_id) FROM listening_profiles p LEFT JOIN listens l ON l.profile_id=p.id GROUP BY p.id"))
        assert actual == expected, (actual, expected)
        assert conn.execute("SELECT max(version_id) FROM goose_db_version WHERE is_applied=1").fetchone() == (3,)
        assert conn.execute("PRAGMA integrity_check").fetchall() == [("ok",)]
        assert conn.execute("PRAGMA foreign_key_check").fetchall() == []
        assert conn.execute("SELECT count(*) FROM listens WHERE track_id=1 AND listened_at=?", (BASE,)).fetchone() == (2,)
        assert conn.execute("SELECT count(*) FROM listens WHERE track_id=3").fetchone() == (0,)
        if not large:
            tracks = conn.execute("SELECT profile_id,track_id,count(*) FROM listens GROUP BY profile_id,track_id ORDER BY 1,2").fetchall()
            assert tracks == [(1, 1, 2), (1, 4, 1), (2, 1, 1), (2, 2, 1)]
            durations = conn.execute("SELECT l.profile_id,sum(t.duration) FROM listens l JOIN tracks t ON t.id=l.track_id GROUP BY l.profile_id ORDER BY l.profile_id").fetchall()
            assert durations == [(1, 560), (2, 180)]
            albums = conn.execute("SELECT l.profile_id,t.release_id,count(*) FROM listens l JOIN tracks t ON t.id=l.track_id GROUP BY 1,2 ORDER BY 1,2").fetchall()
            assert albums == [(1, 1, 3), (2, 1, 1), (2, 2, 1)]
            artists = conn.execute("SELECT l.profile_id,a.artist_id,count(*) FROM listens l JOIN artist_tracks a ON a.track_id=l.track_id GROUP BY 1,2 ORDER BY 1,2").fetchall()
            assert artists == [(1, 1, 3), (1, 2, 2), (2, 1, 1), (2, 2, 2)]
        try:
            conn.execute("INSERT INTO listens(profile_id,track_id,listened_at) VALUES(1,1,?)", (BASE,))
        except sqlite3.IntegrityError:
            conn.rollback()
        else:
            raise AssertionError("same-profile duplicate was accepted")
        if large:
            rows = conn.execute("SELECT listened_at,track_id FROM listens WHERE profile_id=1 ORDER BY listened_at,track_id").fetchall()
            assert len(rows) == 1005 and rows[999][0] == rows[1000][0]
        return actual


def main():
    """Verify both variants without leaving a database on disk."""
    results = {}
    for large in (False, True):
        with disposable_fixture(large) as path:
            results["large" if large else "small"] = verify(path, large)
    print(json.dumps({"reference_now_not_injected": FIXED_NOW, "reference_timezone": PROCESS_TZ,
                      "sqlite": sqlite3.sqlite_version, "results": results}, indent=2))


if __name__ == "__main__":
    main()
