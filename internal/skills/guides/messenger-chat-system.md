# Messenger / Chat System Design Guide

Design a messaging system like WhatsApp or Facebook Messenger that supports real-time 1:1 chat, delivery status tracking, and user presence at scale.

## Functional Requirements
- Users can send and receive 1:1 messages with sent/delivered/read status
- Users can manage contacts (search, add/remove, block) and start chats
- Users can use multiple devices with synced history and message state
- Users can compose/send offline with automatic retry on reconnect

## Non-Functional Requirements
- Low end-to-end latency for real-time delivery (target <200-300ms)
- High availability for chat, presence, and reconnect scenarios
- Strong per-conversation ordering with idempotent retries
- Scalable storage for long-lived history and compliance operations

## Relevant Patterns

**Real-time Updates**: Bi-directional low-latency updates for messages, presence, typing, and receipts through WebSockets + pub/sub.

**Scaling Writes**: Bursty message writes, retries, and acknowledgments require sharding, batching, and log-based ingestion.

**Scaling Reads**: Conversation history is read frequently across devices; use caching, pagination, and hot-partition mitigation.

## Common Deep Dives

**How do you ensure ordering, deduplication, and delivery guarantees across devices and flaky networks?**
- Use server-assigned per-conversation sequence numbers
- Use idempotent message IDs so retries do not duplicate messages
- Track per-device checkpoints/high-watermarks
- Backfill from last checkpoint after reconnect
- Shard by `conversationID` and process in-order within each shard
- Resolve conflicts via server sequence and ingestion time

**How do you keep low-latency delivery and accurate presence with millions of concurrent connections?**
- Use regional WebSocket gateways for foreground sessions
- Use push notifications for backgrounded devices
- Route users to nearest region (Geo-DNS/anycast)
- Track session/presence in-memory with heartbeats and timeouts
- Coalesce rapid presence flips to reduce storms
- Add rate limiting, backpressure, and small-message/ack batching

**How do you design multi-device sync so history, drafts, and receipts converge?**
- Assign unique device identities and device-specific cursors
- Sync deltas since last cursor and paginate deep backfills
- Define deterministic conflict rules (for edits/deletes and receipts)
- Aggregate per-device receipts into user-level state
- Support selective sync + compression for bandwidth control
- On device unlink, revoke device keys/cursors and force resync

**How do you keep storage efficient at 1B+ users while preserving fast reads and compliance?**
- Store messages partitioned by `conversationID` with time clustering
- Apply TTL/cold tier/compaction for storage lifecycle
- Support retention policies and delete workflows (soft delete + async purge)
- Keep primary indexes minimal to avoid write/read amplification
- Stream to separate search index if full-text search is required

## Relevant Technologies
- **Kafka**: durable ordered ingestion, decoupled delivery workers, retry/backpressure handling
- **Redis**: presence/session routing, pub/sub fan-out, idempotency keys, rate limiting
- **Cassandra**: wide-row conversation timelines and cursors with scalable write throughput

## Red Flags to Avoid

**Hand-wavy delivery semantics**: No concrete ordering/dedup/retry model leads to duplicates and missing messages.

**Single hot partition or monolithic message table**: Creates hotspots and tail-latency failures under burst.

**Vague multi-device synchronization**: No per-device cursors/conflict rules causes gaps and inconsistent receipts.

## Similar Problems to Practice
- WhatsApp
- FB Live Comments
- Google Docs
