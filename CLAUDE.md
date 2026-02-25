# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o bin/bonk ./cmd/bonk   # Build
./bin/bonk                         # Drill (smart selection: due → new → random)
./bin/bonk ds                      # Data structures only
./bin/bonk algo                    # Algorithm patterns only
./bin/bonk sys                     # System design only
./bin/bonk --skill hash-maps       # Specific skill
./bin/bonk list                    # List all skills
./bin/bonk stats                   # Show progress
./bin/bonk serve                   # Web terminal for mobile access (requires ttyd)
```

## Environment

- `ANTHROPIC_API_KEY` - Required for drills
- `BONK_MODEL` - Optional, defaults to claude-sonnet-4-20250514

## Architecture

Bonk is an LLM-powered spaced repetition CLI for drilling technical skills using Socratic questioning.

**Core flow**: `cmd/bonk/main.go` → selects skill via SM-2 priority → launches `tui/tui.go` Bubble Tea app → `llm/client.go` manages conversation with Claude API → `db/db.go` persists sessions and scheduling.

**Key packages**:
- `internal/skills/skills.go` - Skill registry with `init()` that registers all skills. Each skill has facets (angles to probe) and example problems. Skills are grouped by domain: data-structures, algorithm-patterns, system-design.
- `internal/db/db.go` - SQLite persistence at `~/.bonk/data.sqlite`. Schema has `sessions`, `exchanges`, and `scheduling` tables. Implements SM-2 algorithm for spaced repetition in `FinishSession()`.
- `internal/llm/client.go` - Anthropic API client. `BuildSystemPrompt()` constructs the Socratic coach prompt including skill facets, history context, and difficulty adjustment based on past performance.
- `internal/tui/tui.go` - Bubble Tea TUI with states: `stateLoading` → `stateDrilling` → `stateRating`. Manages conversation exchanges and rating input (1-4 scale).

**SM-2 scheduling**: The scheduler uses SM-2 algorithm (see `FinishSession` in db.go). Rating 1-2 = lapse (reset interval), 3-4 = success (multiply interval by easiness factor). Interval capped at 365 days.

**Skill selection priority** (in `selectSkill`):
1. Due skills (overdue based on scheduling)
2. New skills (never reviewed)
3. Random fallback

## Roadmap

See `ROADMAP.md` for planned features and priorities. When working on this repo:
- Consult ROADMAP.md before starting new features
- Update ROADMAP.md when completing features or identifying new work
- Use priority labels (P0, P1, P2, P3) to guide what to work on next
