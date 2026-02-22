# bonk

LLM-powered spaced repetition for technical skills. Socratic drilling with SM-2 scheduling.

## Install

```bash
go build -o bin/bonk ./cmd/bonk
# Optional: add bin/ to PATH or copy to ~/bin
```

## Run

```bash
./bin/bonk           # Drill (prioritizes due → new → random)
./bin/bonk ds        # Data structures only
./bin/bonk algo      # Algorithm patterns only
./bin/bonk sys       # System design only
./bin/bonk --skill X # Specific skill
./bin/bonk list      # List all skills
./bin/bonk stats     # Show progress
```

## Config

State: `~/.bonk/data.sqlite`

- `ANTHROPIC_API_KEY` (required for drills)
- `BONK_MODEL` (optional, default: claude-sonnet-4-20250514)
