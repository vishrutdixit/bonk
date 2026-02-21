# infinite-anki (CLI/TUI MVP)

Local-first, interview-style spaced repetition with Socratic follow-ups.

## Install (dev)

```bash
cd apps/infinite-anki-cli
python3 -m venv .venv
source .venv/bin/activate
pip install -U pip
pip install -e .
```

## Run

```bash
infinite-anki
```

## Config

By default, state is stored at `~/.infinite_anki/data.sqlite`.

Provider keys:
- `OPENAI_API_KEY` (optional)
- `ANTHROPIC_API_KEY` (optional)

For the MVP demo, the app works even without keys (uses built-in prompts).
