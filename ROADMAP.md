# Roadmap

Ideas for future development, roughly prioritized.

## High Priority

### CI/CD (S)

GitHub Actions for automated builds and releases.

- Build + test on every push to main
- Release automation: tag → build binaries (darwin/linux, amd64/arm64) → GitHub release
- Claude PR reviews via GitHub integration (optional: auto-review, or @mention to trigger)

### Dev Mode (S)

Developer tooling for debugging and skill iteration.

- `--dev` flag for verbose logging, prompt inspection
- Press `S` during drill to show: skill ID, current facets, difficulty level, session history context
- Useful for tuning skill definitions and prompt behavior

### Onboarding (M)

Initial calibration for new users.

- First-run questionnaire: experience level (beginner/intermediate/advanced), familiar domains
- Option to mark skills as "already know" to skip or reduce frequency
- Seed initial difficulty based on responses instead of waiting for 3 sessions
- Store in `~/.bonk/profile.json`

## Medium Priority

### Better Analytics (M)

Improve `bonk stats` with more actionable insights.

- Per-skill breakdown: rating trend, sessions count, last drilled
- Per-facet tracking: which angles within a skill are weak
- Weak spot detection: "You struggle with X facet of Y skill"
- Visual progress (sparklines or simple ASCII charts)

### Session History (M)

Review past drill sessions.

- `bonk history` command to list recent sessions
- `bonk history <session-id>` to replay Q&A transcript
- Useful for spaced repetition review and self-assessment

### Skill Dependencies (M)

Some skills build on others.

- Define prerequisites (e.g., graph-algorithms needs graphs, bfs, dfs)
- Warn if drilling advanced skill without prerequisite mastery
- Suggest learning path for new users

### LeetCode Practice Suggestions (M)

Surface relevant LeetCode problems based on weak areas.

- Add LeetCode URLs to skill `ExampleProblems`
- `bonk practice` command that suggests 2-3 problems based on recent struggles
- Could show suggestion after a rough drill session

```
$ bonk practice

Based on recent sessions:

  Heaps (struggled with heap property)
  → Kth Largest Element in Array  https://leetcode.com/problems/kth-largest-element-in-an-array/
  → Top K Frequent Elements       https://leetcode.com/problems/top-k-frequent-elements/

  Binary Search (struggled with invariants)
  → Search in Rotated Array       https://leetcode.com/problems/search-in-rotated-sorted-array/
```

## Lower Priority

### Voice Mode (M-L)

Practice explaining concepts out loud, mimicking real interviews.

- **TTS (text-to-speech):** macOS `say` command for questions (easy win), or cloud TTS for better quality
- **STT (speech-to-text):** Whisper (local, private) or macOS Dictation for capturing answers
- **Flow:** Question spoken → user explains aloud → transcribed → sent to LLM → response spoken
- **Modes:** `bonk --voice` for full voice I/O, or hybrid (voice questions, typed answers)
- **Value:** Practices articulation, catches filler words, mimics phone screens

Start simple with TTS-only (`say` on macOS), then add STT later.

### Visualization (M)

Visual aids for tree/graph/DP problems.

- **ASCII art in terminal:** No dependencies, works everywhere, LLM can generate inline
- **Graphviz → image:** Beautiful graphs/trees, opens external viewer
- **Sixel/iTerm2 inline images:** Images in terminal (limited terminal support)
- **Structured data rendering:** LLM outputs structured format, we render it

Use cases:
- Tree structure when drilling BST traversal
- Graph visualization for BFS/DFS
- Linked list diagrams for pointer manipulation
- DP tables for dynamic programming

### Mobile Access (M)

Use bonk from phone.

- **Web terminal (ttyd/gotty):** `ttyd bonk` exposes TUI in browser. Minimal code changes, works immediately.
- **Simple web UI:** Extract core logic into API, build lightweight mobile-friendly frontend.
- **PWA (long term):** Installable web app with offline support, push notifications for daily reminders.

Pragmatic path: Start with ttyd for instant mobile access, iterate toward proper web UI later.

### Streaming LLM Responses (M)

Better UX with real-time response rendering.

- Requires SSE parsing from Anthropic API
- Show tokens as they arrive instead of waiting for full response

### Custom Skills (S)

Allow users to define their own skills.

- YAML/JSON files in `~/.bonk/skills/`
- Same structure as built-in skills (facets, example problems)
- Useful for domain-specific drilling (e.g., company-specific system design)

### Skill Info Command (S)

Quick reference for skill details.

- `bonk info <skill>` shows facets, example problems, description
- `bonk info --all` dumps full skill catalog

### Export to Anki (M)

Export drill content to Anki for offline flashcard practice.

- Generate cards from skill facets
- Include example problems as card prompts

## LC Domain & Archetypes

### Problem Archetypes (M)

A layer above skills — composite patterns that combine multiple skills.

- Archetype = recognizable problem template (e.g., "Cooldown Scheduling")
- Maps to multiple underlying skills (e.g., heaps + greedy + queue)
- Drill focuses on recognition: "Given this problem, what's the archetype?"
- Each archetype has: pattern description, key insight, common variations, representative problems

Example archetypes:
- Cooldown Scheduling (Task Scheduler, Rearrange String K Apart)
- Two Heaps for Median (Find Median from Data Stream)
- Monotonic Stack Optimization (Largest Rectangle, Trapping Rain Water)
- Sliding Window + Hash (Minimum Window Substring)
- BFS with Complex State (Open the Lock, Word Ladder)

### LC Domain (M-L)

New domain for drilling LeetCode problem-solving strategy (not implementation).

```
[lc]
  cooldown-scheduling      Cooldown Scheduling Problems
  sliding-window-hash      Sliding Window + Hash Map
  two-heaps-median         Two Heaps for Running Median
  monotonic-optimization   Monotonic Stack Optimization
  dp-on-intervals          DP on Intervals
  graph-state-bfs          BFS with Complex State
```

Each LC skill drills:
- Pattern recognition: "What's the key insight?"
- Strategy: "What data structures? Why?"
- Complexity: "Time/space?"
- NOT implementation — that's what solving LC is for

### LC Scraper (M)

Build tooling to extract problems and editorials from LeetCode.

- Scrape problem descriptions, constraints, examples
- Extract editorial solutions and explanations
- Map problems to archetypes/skills automatically (or with LLM assist)
- Build up LC domain skill bank from real data
- Could also extract problem tags and difficulty for metadata

## Future Ideas (Unprioritized)

### Interview Simulation Mode

Timed sessions with stricter evaluation.

- 45-minute timed drill
- No hints, stricter Socratic questioning
- Score/feedback at end
- "Phone screen" vs "onsite" modes

### Drill Modes

Different session lengths for different needs.

- Quick drill (5 min) — one focused question
- Standard (15 min) — current default
- Deep dive (30 min) — thorough exploration of a skill

### Gamification

Motivation through streaks and achievements.

- Daily streaks with visual indicator
- Achievements: "10 skills mastered", "7-day streak", "conquered hard mode"
- Optional leaderboards (compare with friends)

### Import LeetCode History

Seed skill ratings from existing LC progress.

- Parse LC submission history
- Map problems to skills
- Bootstrap difficulty calibration

### Code Execution

Actually run code during drills.

- Write and test code snippets
- Language-specific modes (Python/Go/Java idioms)
- Integrate with local compiler/interpreter

## Tech Debt / Fixes

- Fix `struggled` tracking (currently hardcoded to false in db.go)
- Consider splitting large skills.go into per-domain files
- Add unit tests for SM-2 scheduling logic
