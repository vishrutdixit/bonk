from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime, timedelta, timezone


@dataclass
class SchedState:
    stability: float = 1.0  # days
    difficulty: float = 5.0  # 1..10
    lapses: int = 0


def clamp(x: float, lo: float, hi: float) -> float:
    return max(lo, min(hi, x))


def next_due(now: datetime, state: SchedState, rating: int) -> tuple[datetime, SchedState]:
    # rating: 1 Again, 2 Hard, 3 Good, 4 Easy
    s, d, l = state.stability, state.difficulty, state.lapses

    if rating == 1:
        l += 1
        s = clamp(s * 0.5, 0.2, 365)
        d = clamp(d + 0.6, 1, 10)
        return now + timedelta(minutes=20), SchedState(stability=s, difficulty=d, lapses=l)

    if rating == 2:
        s = clamp(s * 1.15, 0.2, 365)
        d = clamp(d + 0.15, 1, 10)
    elif rating == 3:
        s = clamp(s * 1.35, 0.2, 365)
        d = clamp(d - 0.1, 1, 10)
    elif rating == 4:
        s = clamp(s * 1.6, 0.2, 365)
        d = clamp(d - 0.25, 1, 10)

    interval_days = clamp(s * (11 - d) / 10, 0.2, 180)
    return now + timedelta(days=interval_days), SchedState(stability=s, difficulty=d, lapses=l)


def utcnow() -> datetime:
    return datetime.now(timezone.utc)
