# Contributing

Thanks for contributing to `bonk`.

## Prerequisites

- Go (matching `go.mod`)
- `ANTHROPIC_API_KEY` for live drill sessions
- Optional: `ttyd` for `bonk serve`

## Build, Run, Test

```bash
make install-hooks

go build -o bin/bonk ./cmd/bonk

./bin/bonk
./bin/bonk ds
./bin/bonk algo
./bin/bonk sys
./bin/bonk lc
./bin/bonk --skill hash-maps
./bin/bonk list
./bin/bonk info hash-maps
./bin/bonk serve

go test ./...
```

Notes:
- `make install-hooks` sets `core.hooksPath` to `.githooks`.
- Pre-commit hook runs `gofmt -w cmd internal` and `go test ./...`.

## Environment

- `ANTHROPIC_API_KEY` is required for drill sessions.
- `BONK_MODEL` is optional and defaults to `claude-sonnet-4-20250514`.
- Persistent state is stored at `~/.bonk/data.sqlite`.

## Architecture

- `cmd/bonk/main.go`: CLI commands (`drill`, `list`, `info`, `serve`) and skill selection.
- `internal/tui/tui.go`: Bubble Tea state machine and drill UX.
- `internal/llm/client.go`: Anthropic client, prompt construction, response metadata parsing.
- `internal/db/db.go`: SQLite schema, session/exchange persistence, SM-2 scheduling, stats queries.
- `internal/skills/skills.go`: in-code skill catalog and domain mappings.
- `internal/serve/serve.go`: `ttyd` wrapper for phone/web terminal access.

## Scheduling Notes (SM-2)

- Ratings `1-2` are treated as lapse/reset.
- Ratings `3-4` increase interval by easiness factor.
- Interval is capped at 365 days.

## Before You Open a PR

1. Build and run locally.
2. Run tests: `go test ./...`
3. Keep changes focused and avoid unrelated refactors.
4. Note behavior changes that affect drills, scheduling, or stored progress.

## Working Rules

- Prefer minimal, targeted changes over broad refactors.
- Preserve CLI behavior unless intentionally changing product UX.
- Keep TUI interactions predictable and low-friction.
- Keep scheduling logic deterministic and easy to test.
- Use small helper functions over deeply nested logic.
- Do not revert unrelated local edits.
- Run `go test ./...` after code changes.
