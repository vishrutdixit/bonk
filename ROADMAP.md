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

### Coverage Recovery Program (M-L)

Fix CI coverage gate failures in controlled increments while keeping quality standards high.

Current problem:
- CI enforces total coverage >= 70%
- Current total coverage is ~6.3%
- Most packages are at 0%, so this needs a staged program, not a one-off patch

#### Phase 0: Unblock CI with ratcheting guardrails (S)

Goal: keep quality pressure without freezing all development.

- Story COV-00.1: Baseline coverage artifact
  - Capture and store current total coverage as baseline in CI.
  - Acceptance criteria:
    - CI prints baseline and current coverage in logs.
    - Build fails if coverage regresses below baseline by more than tolerance (e.g., 0.2%).

- Story COV-00.2: Replace hard 70% gate with ratchet gate
  - Enforce "no regression + incremental increase target" until overall target is reached.
  - Acceptance criteria:
    - CI fails on regressions.
    - CI passes if coverage holds or improves.
    - Rule is documented in CONTRIBUTING and CI output.

- Story COV-00.3: Re-enable strict gate criteria once target reached
  - Auto-switch or manual switch back to fixed >=70% when total reaches threshold.
  - Acceptance criteria:
    - CI configuration includes clear switch condition and final fixed threshold.

#### Phase 1: Database coverage first (M)

Goal: cover deterministic scheduling/stats logic for biggest ROI.

- Story COV-01.1: DB test harness
  - Add isolated test helper for temporary DB path and deterministic setup/teardown.
  - Acceptance criteria:
    - Tests do not touch user `~/.bonk/data.sqlite`.
    - Tests are parallel-safe and hermetic.

- Story COV-01.2: Session + SM-2 core tests
  - Test `CreateSession`, `FinishSession`, interval updates, lapses, and cap at 365 days.
  - Acceptance criteria:
    - Rating paths 1-4 covered.
    - Lapse/reset and success-growth behaviors asserted.

- Story COV-01.3: Scheduling query tests
  - Cover `GetDueSkills`, `GetDueCount`, `GetDueThisWeek`, `GetNewSkills`.
  - Acceptance criteria:
    - Edge cases validated (empty DB, all-new, all-due, mixed).

- Story COV-01.4: Stats query tests
  - Cover streaks, recent sessions/ratings, weak facets, averages.
  - Acceptance criteria:
    - Empty and populated-history behavior verified.
    - Date-boundary logic (`today`, consecutive-day streak) asserted.

Target: `internal/db` package >=75% coverage.

#### Phase 2: LLM logic coverage (S-M)

Goal: test pure logic and parsing without real network calls.

- Story COV-02.1: Prompt builder tests
  - Cover domain-specific prompt shaping and perf-context behavior.
  - Acceptance criteria:
    - Prompt metadata/structure expectations asserted.

- Story COV-02.2: Response parser tests
  - Cover valid metadata, missing metadata, malformed outputs.
  - Acceptance criteria:
    - Parser degrades gracefully and extracts rating/facet/question type where present.

- Story COV-02.3: Env/model selection tests
  - Cover API key/model env fallback behavior.
  - Acceptance criteria:
    - Defaults and overrides behave deterministically.

Target: `internal/llm` package >=80% coverage (excluding live API call path).

#### Phase 3: TUI deterministic coverage (M)

Goal: cover non-IO-heavy logic and critical state transitions.

- Story COV-03.1: Helper function tests
  - Cover `wordWrap`, `relativeTime`, domain picker cycling, rating glyph/legend mapping.
  - Acceptance criteria:
    - Boundary and formatting edge cases covered.

- Story COV-03.2: Welcome state transition tests
  - Test picker controls (`1-4`, arrows, `j/k`, enter), domain selection persistence.
  - Acceptance criteria:
    - Expected state updates asserted with Bubble Tea key messages.

- Story COV-03.3: Layout regression tests
  - Cover width sync logic to prevent input wrap bugs with sidebar/debug toggles.
  - Acceptance criteria:
    - `syncLayout` behavior asserted for resize and debug sidebar states.

Target: `internal/tui` package >=60% on deterministic functions.

#### Phase 4: CLI + selection behavior (S-M)

Goal: lock down user-facing selection policy.

- Story COV-04.1: `selectSkill` tests
  - Verify priority order due -> new -> random, with and without domain filters.
  - Acceptance criteria:
    - Deterministic tests with seeded/random-controlled data.

- Story COV-04.2: Command surface smoke tests
  - Validate key command wiring (`list`, `info`, drill arg parsing) without live LLM.
  - Acceptance criteria:
    - Basic command parsing and error paths covered.

Target: `cmd/bonk` package >=50% coverage.

#### Phase 5: Tighten policy and make it durable (S)

- Story COV-05.1: Per-package minimums in CI
  - Add package thresholds to prevent one high-coverage package hiding others.
  - Acceptance criteria:
    - CI reports pass/fail per package minimum.

- Story COV-05.2: Coverage reporting UX
  - Add PR-friendly summary (total delta + package deltas).
  - Acceptance criteria:
    - CI output clearly states what changed and why failures happened.

- Story COV-05.3: Finalize fixed global threshold
  - Return to global >=70% once suite stabilizes.
  - Acceptance criteria:
    - Ratchet mode removed or disabled.
    - CONTRIBUTING docs updated with final rule.

## P1: High Priority

Core features that significantly expand capability.

### Mobile Access (S-M)

Use bonk from phone.

- **Phase 1:** Web terminal via `ttyd` — `bonk serve` wraps ttyd, prints local IP for phone access.
- **Phase 1.5:** Tailscale integration — detect Tailscale IP for anywhere-access without port forwarding.
- **Phase 2:** Simple web UI — extract core logic into API, build lightweight mobile-friendly frontend.
- **Phase 3:** PWA — installable web app with offline support, push notifications for daily reminders.

Start with ttyd for instant mobile access (same network). Add Tailscale detection later for remote access.

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

Start with TTS-only (`say` on macOS), add STT via whisper.cpp later.

## P2: Medium Priority

Improve core experience.

### Onboarding (M)

Initial calibration for new users.

- First-run questionnaire: experience level (beginner/intermediate/advanced), familiar domains
- Option to mark skills as "already know" to skip or reduce frequency
- Seed initial difficulty based on responses instead of waiting for 3 sessions
- Store in `~/.bonk/profile.json`

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

### Export to Anki (M)

Export drill content to Anki for offline flashcard practice.

- Generate cards from skill facets
- Include example problems as card prompts

## Future Ideas

Unprioritized explorations.

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
