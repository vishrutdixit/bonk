from __future__ import annotations

import json

from .db import DB
from .scheduler import utcnow


def seed(db: DB) -> None:
    # Small demo seed; later: import Blind 75 -> patterns -> skills.
    skills = [
        {
            "id": "graphs-directed-cycle",
            "title": "Directed graph feasibility (cycle detection)",
            "pattern": "graphs",
            "description": "Given prerequisites (u->v), can you finish all tasks?",
            "rubric": {
                "mustMentionAny": ["cycle", "dag", "topological", "kahn", "indegree", "3-color", "recursion stack"],
                "keyProperty": "Finishable iff the directed graph is acyclic (a DAG).",
            },
            "followups": [
                {"kind": "reframe", "q": "Reachability isn’t quite right. When can you NOT finish?"},
                {"kind": "property", "q": "What specific graph property are you checking for?"},
                {"kind": "mechanics", "q": "How do you detect that with DFS? What states do you track?"},
            ],
            "generator": {
                "families": [
                    "Course schedule",
                    "Build system dependencies",
                    "Deadlock detection framing",
                ]
            },
        },
        {
            "id": "binary-search-first-true",
            "title": "Binary search invariant (first true / lower_bound)",
            "pattern": "binary-search",
            "description": "Given a monotonic predicate, find the first index where it becomes true.",
            "rubric": {
                "mustMentionAny": ["invariant", "lo", "hi", "first true", "lower_bound", "monotonic"],
                "keyProperty": "Maintain an invariant about the boundary of false/true and shrink until lo==hi.",
            },
            "followups": [
                {"kind": "invariant", "q": "State the loop invariant in words."},
                {"kind": "edge", "q": "What if all values are false? all true?"},
            ],
            "generator": {"families": ["first >= x", "min capacity", "koko bananas"]},
        },
        {
            "id": "dp-01-knapsack",
            "title": "0/1 knapsack DP (state + transition)",
            "pattern": "dp",
            "description": "Pick items at most once to maximize value under capacity.",
            "rubric": {
                "mustMentionAny": ["dp", "state", "transition", "capacity", "O(nW)"],
                "keyProperty": "DP over items and capacity; 1D optimized needs reverse loop over w.",
            },
            "followups": [
                {"kind": "state", "q": "Define your DP state precisely."},
                {"kind": "transition", "q": "Write the recurrence/transition."},
            ],
            "generator": {"families": ["subset sum", "partition", "knapsack"]},
        },
    ]

    cur = db.conn.cursor()
    for s in skills:
        cur.execute(
            """
            INSERT INTO skills (id, title, pattern, description, rubric_json, followups_json, generator_json)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(id) DO UPDATE SET
              title=excluded.title,
              pattern=excluded.pattern,
              description=excluded.description,
              rubric_json=excluded.rubric_json,
              followups_json=excluded.followups_json,
              generator_json=excluded.generator_json
            """,
            (
                s["id"],
                s["title"],
                s["pattern"],
                s["description"],
                json.dumps(s["rubric"]),
                json.dumps(s["followups"]),
                json.dumps(s["generator"]),
            ),
        )
        cur.execute(
            """
            INSERT INTO scheduling (skill_id, due_at)
            VALUES (?, datetime('now'))
            ON CONFLICT(skill_id) DO NOTHING
            """,
            (s["id"],),
        )

    db.conn.commit()
