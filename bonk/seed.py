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
                {"kind": "mechanics", "q": "How would you detect a cycle with DFS? (e.g., 3-color / recursion stack)"},
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
            "id": "intervals-merge",
            "title": "Merge Intervals (sort + merge condition)",
            "pattern": "intervals",
            "description": "Given intervals, merge all overlapping intervals. Explain approach + complexity.",
            "rubric": {
                "mustMentionAny": ["sort", "start", "merge", "overlap", "n log n"],
                "keyProperty": "Sort by start; overlap if next.start <= cur.end; merge end = max(cur.end, next.end).",
            },
            "followups": [
                {"kind": "mechanics", "q": "What exactly is the overlap condition? (write it as an inequality)"},
                {"kind": "mechanics", "q": "When merging, why do we take max(cur.end, next.end)? What case breaks if we don’t?"},
                {"kind": "edge", "q": "What if one interval is fully contained in another?"},
            ],
            "generator": {
                "families": [
                    "calendar blocks",
                    "ranges in a log",
                    "meeting times",
                ]
            },
        },

        {
            "id": "sliding-window-longest-unique-substring",
            "title": "Longest substring without repeating characters (sliding window)",
            "pattern": "sliding-window",
            "description": "Given a string, find the length of the longest substring without repeating characters.",
            "rubric": {
                "mustMentionAny": ["sliding window", "two pointers", "hash", "set", "map", "O(n)"],
                "keyProperty": "Maintain a window with all unique chars; expand right; if invalid, shrink left (or jump left using last-seen index map).",
            },
            "followups": [
                {"kind": "invariant", "q": "What is the window invariant (what must always be true about the current window)?"},
                {"kind": "mechanics", "q": "What data structure are you tracking, and what’s the update when you see a duplicate?"},
                {"kind": "edge", "q": "Set-based shrink vs last-seen-index jump: what’s the difference?"},
            ],
            "generator": {
                "families": [
                    "longest unique substring",
                    "max length with constraint",
                ]
            },
        },

        {
            "id": "trees-max-depth",
            "title": "Maximum depth of binary tree (DFS recurrence)",
            "pattern": "trees",
            "description": "Given a binary tree, find its maximum depth.",
            "rubric": {
                "mustMentionAny": ["dfs", "recursion", "base case", "1 +", "O(n)"],
                "keyProperty": "Recurrence: depth(node)=0 if null else 1+max(depth(left), depth(right)); time O(n), space O(h).",
            },
            "followups": [
                {"kind": "mechanics", "q": "State the recurrence in one line (including the base case)."},
                {"kind": "edge", "q": "What’s the space complexity, and what does h mean?"},
            ],
            "generator": {
                "families": [
                    "max depth",
                    "height of tree",
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
