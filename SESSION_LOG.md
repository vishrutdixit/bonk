# Bonk Session Log - 2026-02-22

## What Was Built

### SM-2 Spaced Repetition Scheduling (internal/db/db.go)

Implemented proper SM-2 algorithm in `FinishSession`:
- **EF (Easiness Factor)** stored as `difficulty`, starts at 2.5, min 1.3
- **Rating mapping**: 1→1 (Again), 2→2 (Hard), 3→4 (Good), 4→5 (Easy)
- **Lapse handling**: Rating 1-2 resets stability to 1 day, increments lapse counter
- **Interval growth**: `new_interval = stability * EF`, capped at 365 days
- **EF update formula**: `EF' = EF + (0.1 - (5-q) * (0.08 + (5-q) * 0.02))`

Added new DB functions:
- `GetDueSkills()` - Returns skills where `due_at <= now`
- `GetNewSkills(allSkillIDs)` - Returns skill IDs not in scheduling table
- `GetDueCount()` - Returns count of due skills

### Smart Skill Selection (cmd/bonk/main.go)

Added `selectSkill()` function with priority:
1. **Due skills** - Overdue based on SM-2 scheduling
2. **New skills** - Never reviewed (not in scheduling table)
3. **Random** - Fallback when nothing due or new

Updated stats command to show:
```
Skills due: X | New skills: Y
```

### Expanded Skill Set (internal/skills/skills.go)

Added 13 new skills (22 → 35 total):

**Data Structures (2 new):**
- `linked-lists` - Reversal, dummy head, trade-offs vs arrays
- `segment-trees` - Range queries, lazy propagation, when vs prefix sum

**Algorithm Patterns (7 new):**
- `union-find` - Path compression, union by rank, cycle detection
- `fast-slow-pointers` - Floyd's tortoise/hare, cycle start, middle of list
- `prefix-sum` - Range queries, hashmap pattern, 2D prefix sum
- `merge-intervals` - Sort by start, overlap detection, meeting rooms
- `bit-manipulation` - XOR tricks, Brian Kernighan's, bitmask DP
- `shortest-path` - Dijkstra, Bellman-Ford, BFS vs Dijkstra decision

**System Design (4 new):**
- `database-replication` - Leader/follower, sync vs async, failover
- `api-gateway` - Routing, auth at edge, protocol translation
- `cdn` - Edge caching, pull vs push, cache invalidation
- `distributed-coordination` - Leader election, distributed locks, Raft/Paxos basics

### ListIDs Helper (internal/skills/skills.go)

Added `ListIDs() []string` to return all skill IDs for the new skills check.

## Current State

- **Total skills**: 34
- **Total sessions**: 13 (from previous testing)
- **Skills due**: 0 (recent reviews haven't come due yet)
- **New skills**: 25 (not yet reviewed)

## Commands

```bash
./bin/bonk           # Drill (prioritizes due → new → random)
./bin/bonk ds        # Drill data structures only
./bin/bonk algo      # Drill algorithm patterns only
./bin/bonk sys       # Drill system design only
./bin/bonk --skill X # Drill specific skill
./bin/bonk list      # List all skills
./bin/bonk stats     # Show progress
```

## Key Files Modified

1. `internal/db/db.go` - SM-2 algorithm, GetDueSkills, GetNewSkills, GetDueCount
2. `cmd/bonk/main.go` - selectSkill function, stats due count
3. `internal/skills/skills.go` - 13 new skills, ListIDs helper

## Remaining Work (from plan)

- Weak facet targeting in prompts
- End-of-session reflection (LLM suggests skill gaps)
- Question bank export
- Config file for API key
- Install to PATH
