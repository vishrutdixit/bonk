from __future__ import annotations

import os
from pathlib import Path


def data_dir() -> Path:
    p = Path(os.path.expanduser("~/.infinite_anki"))
    p.mkdir(parents=True, exist_ok=True)
    return p


def db_path() -> Path:
    return data_dir() / "data.sqlite"


def config_path() -> Path:
    return data_dir() / "config.json"
