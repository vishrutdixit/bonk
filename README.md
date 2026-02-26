<h1 align="center">bonk</h1>

<p align="center"><strong>Socratic interview drills with spaced repetition</strong></p>

<p align="center">
  <a href="https://github.com/vishrutdixit/bonk/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/vishrutdixit/bonk/actions/workflows/ci.yml/badge.svg?branch=main"></a>
  <a href="https://go.dev/"><img alt="Go Version" src="https://img.shields.io/badge/go-1.25.5-00ADD8?logo=go"></a>
  <a href="https://github.com/vishrutdixit/bonk/releases"><img alt="Latest Release" src="https://img.shields.io/github/v/release/vishrutdixit/bonk"></a>
  <a href="https://github.com/vishrutdixit/bonk/stargazers"><img alt="GitHub Stars" src="https://img.shields.io/github/stars/vishrutdixit/bonk?style=social"></a>
</p>

`bonk` is a terminal app for technical interview prep. It asks probing follow-ups, adapts to your answers, and schedules reviews with SM-2 so you revisit skills at the right time.

## Demo

TODO: add terminal demo GIF and video.

<!--
When assets are ready, replace this block with:

[![bonk demo](dist/demo/bonk-demo.gif)](dist/demo/bonk-demo.mp4)

Suggested paths:
- GIF: dist/demo/bonk-demo.gif
- MP4: dist/demo/bonk-demo.mp4
-->

## Installation

Quick install:

```bash
curl -fsSL https://raw.githubusercontent.com/vishrutdixit/bonk/main/install.sh | bash
```

Manual build:

```bash
go build -o bin/bonk ./cmd/bonk
```

Set your API key and run:

```bash
export ANTHROPIC_API_KEY=your_key_here
./bin/bonk
```

Homebrew: coming soon.

## Why Bonk

- Infinite, self-improving deck of concepts to drill
- Conversation-first practice instead of flashcard memorization
- Smart next-skill selection: due -> new -> random

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
