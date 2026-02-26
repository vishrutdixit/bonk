# bonk

Socratic interview drills with spaced repetition.

`bonk` is a terminal app for technical interview prep. It asks probing follow-ups, adapts to your answers, and schedules review with SM-2 so you revisit skills at the right time.

## Install

```bash
go build -o bin/bonk ./cmd/bonk
```

Set your API key:

```bash
export ANTHROPIC_API_KEY=your_key_here
```

Then run:

```bash
./bin/bonk
```

## Why bonk

- Conversation-first practice instead of flashcard memorization
- Smart next-skill selection: due -> new -> random
- Built-in progress signal on the welcome screen
- Local persistence in SQLite (`~/.bonk/data.sqlite`)

## Common Commands

```bash
./bin/bonk                 # Start drilling (recommended)
./bin/bonk ds              # Data structures only
./bin/bonk algo            # Algorithm patterns only
./bin/bonk sys             # System design only
./bin/bonk lc              # LeetCode patterns only
./bin/bonk --skill hash-maps
./bin/bonk list
./bin/bonk info hash-maps
./bin/bonk --dev           # Debug panel (press S in-session)
```

## Mobile / Remote Drill

```bash
# Install ttyd first
brew install ttyd          # macOS
apt install ttyd           # Linux

./bin/bonk serve
./bin/bonk serve --port 9000
```

Open the printed network URL from your phone on the same WiFi.

## Configuration

- `ANTHROPIC_API_KEY` (required)
- `BONK_MODEL` (optional, defaults to `claude-sonnet-4-20250514`)

## For Contributors

See [CONTRIBUTING.md](/Users/vjd/dev/bonk/CONTRIBUTING.md).
