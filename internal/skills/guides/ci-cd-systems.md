# CI/CD Systems Design Guide

Design a CI/CD platform like GitHub Actions that triggers workflow execution on code push.

## Scale
- 10 million repositories
- 10 pushes/second average, with burst traffic
- Real-time monitoring required

## Functional Requirements
- Workflows trigger automatically on git push using repo-defined workflow files
- Workflows execute in isolated, sandboxed environments with access to permitted secrets only
- Real-time observation: overall status, per-step status, live logs
- Users can view history, retry/cancel runs, access final artifacts

## Core Concepts to Probe

**Multi-Step Workflow Orchestration**: Workflows model as DAGs with job dependencies, retries, and cancellations. Requires durable state machines and versioned execution specs.

**Long-Running Asynchronous Tasks**: Build jobs run for extended periods. Requires async execution, heartbeats, timeouts, and resumable state across worker failures.

**Real-Time Updates at Scale**: Live status and log streaming demand push-based mechanisms (WebSockets/SSE), fan-out capabilities, and backpressure handling.

## Red Flags to Watch For

**Synchronous Processing Without Buffering**: Processing webhooks synchronously fails under bursts and risks losing triggers. Need immediate ack with durable enqueueing.

**Ignoring Multi-Tenant Isolation**: Single repos monopolizing runner fleet causes starvation. Per-tenant quotas and fair scheduling are essential.

**Missing Durable State & Idempotency**: Ad-hoc queuing without persisted state leads to double-execution and lost updates. State machines with idempotent transitions are critical.

## Deep Dives

**Event Ingestion Under Bursts**: How reliably ingest push events during spikes and guarantee exactly-once workflow triggering?
- API gateways with auth/rate limiting
- Idempotency keys from (repo_id, commit_sha, workflow_path)
- Durable logs like Kafka partitioned by repo for ordering

**Workflow Orchestration with Durability**: How orchestrate multi-step DAGs with retries, timeouts, cancellations?
- Model runs/steps as transactional state machines
- Exponential backoff per-step
- Separate orchestration from execution
- Version DAG specs with the commit

**Multi-Tenant Isolation & Security**: How prevent noisy neighbors and secure secrets?
- Per-tenant quotas/concurrency limits
- Fair scheduling algorithms
- Sandbox jobs in containers/VMs with ephemeral credentials
- Scope tokens narrowly with short TTLs

**Real-Time Logs at Scale**: How stream logs without overwhelming storage?
- Stream step logs as frames from runners
- Fan-out via pub/sub to WebSocket/SSE clients
- Per-connection backpressure
- Log size caps, compression, TTLs, tiered retention

## Key Technologies
- **Kafka**: Smooths bursty traffic, durable ordering per repo/tenant, retries and dead-letter handling
- **Redis**: Ephemeral coordination—heartbeats, distributed locks, caching, rate limiting, log fan-out
- **PostgreSQL**: Durable control plane for runs, steps, DAG definitions, idempotency keys; shardable by tenant

## Related Problems
- Job Scheduler (DAGs, workers, heartbeats)
- Facebook Live Comments (real-time high-fan-out)
- Distributed Rate Limiter (per-tenant quotas, fairness)
