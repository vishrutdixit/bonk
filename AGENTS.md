# AGENTS.md

LLM guidance for working in `/Users/vjd/dev/bonk`.

## Project Summary

Bonk is an LLM-powered spaced repetition CLI for technical interview prep. It drills users with Socratic questioning, stores progress in SQLite, and schedules reviews with SM-2.

Core user flow:
1. CLI selects next skill with priority: due -> new -> random.
2. Bubble Tea TUI runs a coach/user drill conversation.
3. Anthropic API generates questions and feedback.
4. Session rating updates scheduling state for the next review.

## Build, Run, Test

```bash
go build -o bin/bonk ./cmd/bonk
./bin/bonk
./bin/bonk ds
./bin/bonk algo
./bin/bonk sys
./bin/bonk lc
./bin/bonk --skill hash-maps
./bin/bonk list
./bin/bonk stats
./bin/bonk serve
go test ./...
```

## Environment

- `ANTHROPIC_API_KEY` is required for drill sessions.
- `BONK_MODEL` is optional and defaults to `claude-sonnet-4-20250514`.
- Persistent state is stored at `~/.bonk/data.sqlite`.

## Architecture

- `cmd/bonk/main.go`: CLI commands (`drill`, `list`, `stats`, `serve`) and skill selection.
- `internal/tui/tui.go`: Bubble Tea state machine and drill UX.
- `internal/llm/client.go`: Anthropic client, prompt construction, response metadata parsing.
- `internal/db/db.go`: SQLite schema, session/exchange persistence, SM-2 scheduling, stats queries.
- `internal/skills/skills.go`: in-code skill catalog and domain mappings.
- `internal/serve/serve.go`: `ttyd` wrapper for phone/web terminal access.

SM-2 note:
- Ratings `1-2` are treated as lapse/reset.
- Ratings `3-4` increase interval by easiness factor.
- Interval is capped at 365 days.

## Working Rules

- Start by checking `git status --short`, `README.md`, and relevant files.
- Prefer minimal, targeted changes over broad refactors.
- Preserve CLI behavior unless explicitly asked to change it.
- Keep TUI interactions predictable and low-friction.
- Keep scheduling logic deterministic and easy to test.
- Use small helper functions over deeply nested logic.
- Do not revert unrelated local edits.
- Run `go test ./...` after code changes.

## Safety Rules

- Never run destructive git/file commands unless explicitly requested.
- Flag assumptions and behavior-risk changes, especially anything that can alter progress data.

## Project-Specific Checks

- Prompt changes: ensure output metadata still matches parser expectations.
- Drill-flow changes: validate turn counting, state transitions, and rating submission.
- Skill-catalog changes: keep domain mappings and list output consistent.
- Stats/scheduling changes: verify query semantics and edge cases for empty histories.

## Planning Guidance

- Consult `ROADMAP.md` before starting net-new features.
- Update `ROADMAP.md` when completing roadmap items or identifying major new work.
- Use roadmap priorities (P0, P1, P2, P3) to guide what to build next.
