from __future__ import annotations

import sqlite3
from dataclasses import dataclass
from pathlib import Path

from .paths import db_path


SCHEMA = """
CREATE TABLE IF NOT EXISTS skills (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  pattern TEXT NOT NULL,
  description TEXT NOT NULL,
  rubric_json TEXT NOT NULL,
  followups_json TEXT NOT NULL,
  generator_json TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS scheduling (
  skill_id TEXT PRIMARY KEY,
  due_at TEXT NOT NULL,
  stability REAL NOT NULL DEFAULT 1,
  difficulty REAL NOT NULL DEFAULT 5,
  lapses INTEGER NOT NULL DEFAULT 0,
  last_rating INTEGER,
  last_reviewed_at TEXT,
  FOREIGN KEY(skill_id) REFERENCES skills(id)
);

CREATE TABLE IF NOT EXISTS reviews (
  id TEXT PRIMARY KEY,
  skill_id TEXT NOT NULL,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  prompt TEXT NOT NULL,
  answer1 TEXT,
  followups_asked_json TEXT NOT NULL DEFAULT '[]',
  answer2 TEXT,
  key_property_revealed TEXT,
  rating INTEGER,
  failure_mode TEXT,
  FOREIGN KEY(skill_id) REFERENCES skills(id)
);
"""


@dataclass
class DB:
    conn: sqlite3.Connection


def open_db(path: Path | None = None) -> DB:
    p = path or db_path()
    conn = sqlite3.connect(str(p))
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA journal_mode=WAL")
    conn.executescript(SCHEMA)
    return DB(conn)
