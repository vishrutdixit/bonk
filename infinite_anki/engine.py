from __future__ import annotations

import json
import time
import uuid
from dataclasses import dataclass
from datetime import datetime

from .db import DB
from .scheduler import SchedState, next_due, utcnow
from .seed import seed


@dataclass
class Skill:
    id: str
    title: str
    pattern: str
    description: str
    rubric: dict
    followups: list[dict]


@dataclass
class ReviewSession:
    review_id: str
    skill: Skill
    prompt: str
    started_at: float
    key_property: str


def ensure_seed(db: DB) -> None:
    row = db.conn.execute("SELECT COUNT(*) AS n FROM skills").fetchone()
    if int(row["n"]) == 0:
        seed(db)


def due_skills(db: DB, limit: int = 50) -> list[Skill]:
    ensure_seed(db)
    rows = db.conn.execute(
        """
        SELECT s.*, sch.due_at
        FROM skills s
        JOIN scheduling sch ON sch.skill_id = s.id
        WHERE datetime(sch.due_at) <= datetime('now')
        ORDER BY datetime(sch.due_at) ASC
        LIMIT ?
        """,
        (limit,),
    ).fetchall()

    out: list[Skill] = []
    for r in rows:
        out.append(
            Skill(
                id=r["id"],
                title=r["title"],
                pattern=r["pattern"],
                description=r["description"],
                rubric=json.loads(r["rubric_json"]),
                followups=json.loads(r["followups_json"]),
            )
        )
    return out


def start_review(db: DB, skill_id: str) -> ReviewSession:
    r = db.conn.execute("SELECT * FROM skills WHERE id=?", (skill_id,)).fetchone()
    if not r:
        raise ValueError("skill_not_found")

    rubric = json.loads(r["rubric_json"])
    followups = json.loads(r["followups_json"])

    skill = Skill(
        id=r["id"],
        title=r["title"],
        pattern=r["pattern"],
        description=r["description"],
        rubric=rubric,
        followups=followups,
    )

    prompt = skill.description
    review_id = uuid.uuid4().hex
    db.conn.execute(
        "INSERT INTO reviews (id, skill_id, started_at, prompt) VALUES (?, ?, datetime('now'), ?)",
        (review_id, skill.id, prompt),
    )
    db.conn.commit()

    return ReviewSession(
        review_id=review_id,
        skill=skill,
        prompt=prompt,
        started_at=time.time(),
        key_property=str(rubric.get("keyProperty", "")),
    )


def pick_followup(skill: Skill, answer: str) -> tuple[str | None, str | None]:
    a = answer.lower()
    must = [str(x).lower() for x in skill.rubric.get("mustMentionAny", [])]
    hit = any(k in a for k in must)

    if not hit:
        # prioritize reframe/property
        for kind in ("reframe", "property"):
            for f in skill.followups:
                if f.get("kind") == kind:
                    return f.get("q"), "missing_key_concept"
        return (skill.followups[0].get("q") if skill.followups else None), "missing_key_concept"

    # mechanics/edge
    for kind in ("mechanics", "edge", "invariant", "state", "transition"):
        for f in skill.followups:
            if f.get("kind") == kind:
                return f.get("q"), None
    return (skill.followups[0].get("q") if skill.followups else None), None


def finish_review(db: DB, review_id: str, rating: int, answer2: str | None = None, key_property_revealed: str | None = None) -> str:
    row = db.conn.execute(
        "SELECT skill_id FROM reviews WHERE id=?", (review_id,)
    ).fetchone()
    if not row:
        raise ValueError("review_not_found")
    skill_id = row["skill_id"]

    sch = db.conn.execute(
        "SELECT stability, difficulty, lapses FROM scheduling WHERE skill_id=?",
        (skill_id,),
    ).fetchone()
    state = SchedState(
        stability=float(sch["stability"]),
        difficulty=float(sch["difficulty"]),
        lapses=int(sch["lapses"]),
    )

    now = utcnow()
    due_at, new_state = next_due(now, state, rating)

    db.conn.execute(
        """
        UPDATE reviews
        SET finished_at=datetime('now'), rating=?, answer2=?, key_property_revealed=?
        WHERE id=?
        """,
        (rating, answer2, key_property_revealed, review_id),
    )
    db.conn.execute(
        """
        UPDATE scheduling
        SET due_at=?, stability=?, difficulty=?, lapses=?, last_rating=?, last_reviewed_at=?
        WHERE skill_id=?
        """,
        (
            due_at.isoformat(),
            new_state.stability,
            new_state.difficulty,
            new_state.lapses,
            rating,
            now.isoformat(),
            skill_id,
        ),
    )
    db.conn.commit()
    return due_at.isoformat()
