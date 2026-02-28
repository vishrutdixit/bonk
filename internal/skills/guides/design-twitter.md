# Design Twitter Guide

Design a social media platform like Twitter/X where users can post tweets, follow other users, and view a personalized home timeline.

## Functional Requirements
- Users can post tweets (text, images, videos) and view their own timeline
- Users can follow/unfollow other users
- Users can view a home timeline of tweets from people they follow
- Users can like, retweet, and reply to tweets

## Non-Functional Requirements
- Low latency for home timeline reads (<200ms)
- High availability (users expect Twitter to always be up)
- Eventual consistency is acceptable for feed (can be slightly stale)
- Scale to 500M+ users, 1B+ tweets/day

## Relevant Patterns

**Fan-out on Write**: Pre-compute timelines when tweets are posted. Good for users with few followers. Bad for celebrities.

**Fan-out on Read**: Compute timeline on request by fetching from followed users. Good for celebrities. Higher read latency.

**Hybrid Approach**: Fan-out on write for normal users, fan-out on read for celebrities (>1M followers).

## Common Deep Dives

**How do you handle celebrity accounts posting to millions of followers?**
- Don't fan-out celebrities' tweets to all follower timelines
- Mark users with >N followers as "celebrities"
- On timeline read, merge pre-computed timeline with celebrity tweets
- Cache celebrity tweets aggressively (high read volume)
- Consider separate "celebrity tweet cache" per region

**How do you generate and rank the home timeline?**
- Fetch candidate tweets from followed users (fan-out + celebrity merge)
- Apply ranking model (engagement prediction, recency, relevance)
- Filter (muted users, blocked, already seen)
- Paginate with cursor-based pagination
- Cache ranked timeline with TTL, invalidate on new activity

**How do you handle the write storm when a celebrity posts?**
- Don't fan-out to followers immediately
- Instead, on read, merge celebrity tweets into user's timeline
- Tweets stored in celebrity's own timeline (not fanned out)
- Use bloom filters or timestamp checks to avoid re-fetching

**How do you implement trending topics?**
- Stream tweets through real-time pipeline (Kafka/Flink)
- Count hashtags/keywords in sliding windows (5m, 1h, 24h)
- Filter for velocity (rate of increase, not just volume)
- Personalize by region/interests
- Anti-gaming: filter bots, spam, coordinated campaigns

## Relevant Technologies
- **Redis**: Timeline caching, user-to-followers mapping, rate limiting
- **Cassandra**: Tweet storage (partition by user_id, cluster by timestamp)
- **Kafka**: Real-time event streaming for fan-out workers, trending
- **CDN**: Media delivery (images, videos)

## Red Flags to Avoid

**Pure fan-out on write**: "Every tweet goes to every follower's timeline" breaks for celebrities.

**No caching strategy**: Timeline reads are extremely hot; must cache aggressively.

**Ignoring the celebrity problem**: Not addressing how to handle 50M+ follower accounts.

## Similar Problems to Practice
- Instagram Feed
- Facebook News Feed
- LinkedIn Feed
- TikTok For You Page
