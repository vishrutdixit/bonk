# Roadmap

## P0: Foundation (Ship First)

Quick wins that unblock other work.

### CI/CD (S)

GitHub Actions for automated builds and releases.

- Build + test on every push to main
- Release automation: tag → build binaries (darwin/linux, amd64/arm64) → GitHub release
- Claude PR reviews via GitHub integration (optional: auto-review, or @mention to trigger)

Status: Phase 2 implemented (February 25, 2026): CI runs formatting/test/build on push/PR, and release workflow publishes darwin/linux binaries on version tags. Claude PR reviews remain optional/pending integration.

### Dev Mode (S)

Developer tooling for debugging and skill iteration.

- `--dev` flag for verbose logging, prompt inspection
- Press `S` during drill to show: skill ID, current facets, difficulty level, session history context
- Useful for tuning skill definitions and prompt behavior

Status: Implemented (February 25, 2026) with `--dev` and in-session `S` debug panel.

### Test Coverage (M-L)

Incrementally improve test coverage across packages. Target 70% overall.

Priority order:
1. `internal/db` — deterministic scheduling/stats logic (highest ROI)
2. `internal/llm` — prompt building and response parsing (no live API calls)
3. `internal/tui` — helper functions and state transitions
4. `cmd/bonk` — skill selection and CLI parsing

## P1: High Priority

Core features that significantly expand capability.

### Mobile Access (S-M)

Use bonk from phone.

- **Phase 1:** Web terminal via `ttyd` — `bonk serve` wraps ttyd, prints local IP for phone access.
- **Phase 1.5:** Tailscale integration — detect Tailscale IP for anywhere-access without port forwarding.
- **Phase 2:** Simple web UI — extract core logic into API, build lightweight mobile-friendly frontend.
- **Phase 3:** PWA — installable web app with offline support, push notifications for daily reminders.

Start with ttyd for instant mobile access (same network). Add Tailscale detection later for remote access.

Status: Phase 1.5 implemented (February 28, 2026). `bonk serve` wraps ttyd, auto-detects Tailscale IP for remote access.

### LC Domain & Archetypes (M-L)

New domain for drilling LeetCode problem-solving strategy (not implementation).

**Problem Archetypes** — a layer above skills, composite patterns:
- Archetype = recognizable problem template (e.g., "Cooldown Scheduling")
- Maps to multiple underlying skills (e.g., heaps + greedy + queue)
- Drill focuses on recognition: "Given this problem, what's the archetype?"

Example archetypes:
- Cooldown Scheduling (Task Scheduler, Rearrange String K Apart)
- Two Heaps for Median (Find Median from Data Stream)
- Monotonic Stack Optimization (Largest Rectangle, Trapping Rain Water)
- Sliding Window + Hash (Minimum Window Substring)
- BFS with Complex State (Open the Lock, Word Ladder)

**LC Domain structure:**
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

### Voice Mode (M)

Practice explaining concepts out loud, mimicking real interviews. **Free with local tools.**

- **TTS:** macOS `say` command (free, built-in) or `piper` (local neural TTS)
- **STT:** `whisper.cpp` (free, runs locally, great accuracy)
- **Flow:** Question spoken → user explains aloud → transcribed → sent to LLM → response spoken
- **Modes:** `bonk --voice` for full voice I/O, or hybrid (voice questions, typed answers)

Status: Implemented (February 26, 2026). TTS via macOS `say` at 280 wpm, STT via whisper.cpp (tiny model). Space to record, 's' to skip speech. Homebrew installs all dependencies automatically.

## P2: Medium Priority

Improve core experience.

### Onboarding (M-L)

Interactive first-run experience that personalizes the entire program.

**What to gather:**
- Experience level (new grad / 2-5 years / senior / staff+)
- Current role (SWE, infra, ML, frontend, etc.)
- Interview timeline (casual prep / interviewing soon / active interviews)
- Focus areas (which domains to prioritize: DS, algo, system design, LC)
- Self-assessed strengths ("I'm good at trees and graphs")
- Self-assessed weaknesses ("I struggle with DP and system design")
- Learning style preferences (more hints vs sink-or-swim, encouraging vs critical)

**How it influences the program:**
- Skill selection: prioritize weak areas, deprioritize strengths
- Difficulty calibration: start harder for experienced users
- Prompt tone: more encouraging for beginners, more critical for senior
- Domain focus: weight skill selection toward chosen domains
- Feedback style: adjust based on learning preferences
- Session length defaults: shorter for casual, longer for intensive prep

**Implementation approaches:**

*Option A: Simple CLI prompts*
- Series of multiple-choice questions on first run
- Quick to implement, works everywhere
- Less conversational, more robotic

*Option B: LLM-powered conversational onboarding*
- Natural conversation: "Tell me about your background..."
- LLM extracts structured data from free-form responses
- More engaging, can ask follow-ups
- Higher cost (LLM calls), more complex

*Option C: Hybrid*
- Start with a few structured questions (experience, timeline)
- Use LLM for open-ended parts (strengths/weaknesses)
- Balance between structure and flexibility

**Storage:**
- `~/.bonk/profile.json` for user profile
- Add `profile` table to SQLite for structured data
- Profile can be updated: `bonk profile` to view/edit

**Profile evolution:**
- Initial onboarding seeds the profile
- Performance data refines it over time (claimed strength but struggling? adjust)
- Periodic check-ins: "You've improved at DP - want to increase difficulty?"

**Open questions:**
- Should profile affect SM-2 scheduling or just skill selection?
- How to handle profile updates without losing calibration data?
- Should there be "interview mode" that ignores profile and goes hard?

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

Status: Partially implemented (February 28, 2026). `bonk review` shows last session transcript, `bonk review --feedback` gets AI analysis of delivery/communication patterns. Still needs: list of sessions, session by ID.

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

## P3: Lower Priority

Nice-to-haves.

### Visualization (M)

Visual aids for tree/graph/DP problems.

- **ASCII art in terminal:** No dependencies, works everywhere, LLM can generate inline
- **Graphviz → image:** Beautiful graphs/trees, opens external viewer
- **Structured data rendering:** LLM outputs structured format, we render it

Use cases: tree traversal, graph BFS/DFS, linked list manipulation, DP tables.

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

Status: Implemented.

### Export to Anki (M)

Export drill content to Anki for offline flashcard practice.

- Generate cards from skill facets
- Include example problems as card prompts

## Future Ideas

Unprioritized explorations.

### System Design Practical Domain

Full interview simulations (Design Twitter, Design Uber, etc.) with phase-aware prompting following the Hello Interview framework.

Phases: Requirements → Core Entities → API Design → Data Flow → High-Level Design → Deep Dives

Status: Implemented (February 28, 2026). `bonk sysp` domain with 6 practical skills, phase tracking in TUI header, extended turn limits (40 turns).

### Interview Simulation Mode

- 45-minute timed drill
- No hints, stricter Socratic questioning
- Score/feedback at end
- "Phone screen" vs "onsite" modes

### Drill Modes

- Quick drill (5 min) — one focused question
- Standard (15 min) — current default
- Deep dive (30 min) — thorough exploration

### Gamification

- Daily streaks with visual indicator
- Achievements: "10 skills mastered", "7-day streak", "conquered hard mode"
- Optional leaderboards (compare with friends)

### Import LeetCode History

- Parse LC submission history
- Map problems to skills
- Bootstrap difficulty calibration

### Code Execution

- Write and test code snippets during drills
- Language-specific modes (Python/Go/Java idioms)
- Integrate with local compiler/interpreter

## Tech Debt

- Fix `struggled` tracking (currently hardcoded to false in db.go)
- Consider splitting large skills.go into per-domain files
- Add unit tests for SM-2 scheduling logic
