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
./bin/bonk --dev     # Drill with dev debug panel (press S in-session)
./bin/bonk ds        # Data structures only
./bin/bonk algo      # Algorithm patterns only
./bin/bonk sys       # System design only
./bin/bonk lc        # LeetCode patterns only
./bin/bonk --skill X # Specific skill
./bin/bonk list      # List all skills
./bin/bonk info X    # Show skill details
```

`bonk` welcome screen now shows progress stats (sessions, due counts, streak, recent ratings).

## Mobile Access

Drill from your phone using web terminal:

```bash
# Install ttyd first
brew install ttyd  # macOS
apt install ttyd   # Linux

# Start web terminal
./bin/bonk serve              # http://localhost:8080
./bin/bonk serve --port 9000  # Custom port
```

Open the "Network" URL from your phone (same WiFi network).

## Config

State: `~/.bonk/data.sqlite`

- `ANTHROPIC_API_KEY` (required for drills)
- `BONK_MODEL` (optional, default: claude-sonnet-4-20250514)
