# Online Chess Platform Design Guide

Design a real-time chess platform similar to Chess.com that supports gameplay across devices, strict move validation, matchmaking, leaderboards, and undo.

## Functional Requirements
- Users can create/join games via quick matchmaking or direct challenges
- Games run in real time with legal move validation and accurate visible clocks
- Users can view global/personal leaderboards and see rating updates shortly after completion
- Users can review/resume games across devices with complete permanent history

## Non-Functional Requirements
- Target sub-150ms end-to-end move latency for good real-time experience
- Strong consistency per game for move order, clocks, and legality
- High availability for active games and reconnect continuity
- Durable event history for replay, auditability, and long-term storage

## Relevant Patterns

**Real-time Updates**: Moves, clocks, and presence changes must propagate quickly between players and spectators using push channels and server authority.

**Dealing with Contention**: Concurrent input from multi-device users and hot resources (queues, leaderboards) require sequencing, idempotency, and safe ownership.

**Scaling Reads**: Leaderboards and game-history views are read-heavy; use caching and precomputed aggregates.

## Common Deep Dives

**How will you achieve sub-150ms move latency and accurate clocks across regions?**
- Use server-authoritative WebSockets and route to a region-near game coordinator
- Keep one authoritative coordinator per game for move order and clock truth
- Drive official clocks on the server and send periodic corrections to clients
- Use heartbeats/PING-PONG for latency measurement and drift handling
- Support fast reconnect via session-resume tokens and sticky routing by `game_id`
- For failover, promote a warm standby with the latest persisted move log

**How will you guarantee move ordering, idempotency, and safe undo on flaky networks?**
- Treat moves as append-only events with monotonic sequence numbers
- Require `expected_seq` in client requests; reject or reconcile out-of-order moves
- Use move UUIDs + conditional writes (CAS/version checks) for idempotency
- Persist all events for deterministic replay on reconnect
- Implement undo as server-mediated negotiation (both players consent)
- Roll back last N events atomically and restore clocks from logged deltas

**How will you design scalable and fair matchmaking while preventing abuse?**
- Maintain queues by time control and rating bands in low-latency in-memory stores
- Gradually widen acceptable rating/geography over wait time to control tail latency
- Keep matching stateless/idempotent; use token claim or lock to avoid double matches
- Handle cancels/timeouts explicitly to avoid stale queue entries
- Add anti-abuse controls for rapid enqueue/dequeue and multi-device duplicates

**How will ratings/leaderboards update quickly while storing games forever?**
- Emit `game_complete` events to a durable log and compute ratings asynchronously
- Update leaderboards in a fast store backed by a durable system of record
- Store moves as event log + periodic snapshots for efficient replay
- Partition archival tables by user/time to maintain index health
- Batch/debounce leaderboard writes and rebuild aggregates from event stream when needed

## Relevant Technologies
- **Redis**: matchmaking queues, presence/session tokens, active game lookup, ZSET leaderboards, rate limiting
- **Kafka**: decouple real-time gameplay from async rating, anti-cheat analytics, and archival pipelines
- **PostgreSQL**: transactional durability for users/games/event logs with partitioning for long-term scale

## Red Flags to Avoid

**Client-authoritative moves/clocks**: Clients must not be source of truth for legality, ordering, or official time.

**Missing ordering/idempotency for moves and undo**: No sequence numbers/acks/idempotency leads to duplicates and corruption.

**Storing full snapshots on every move**: Prefer compact event deltas plus periodic snapshots for efficient storage and replay.

## Similar Problems to Practice
- WhatsApp (1:1 low-latency messaging, ordering, reconnect)
- Google Docs (streamed operations, consistency, undo/redo semantics)
- FB Live Comments (real-time fan-out and ordering under burst)
