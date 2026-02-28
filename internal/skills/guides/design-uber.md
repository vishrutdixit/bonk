# Design Uber Guide

Design a ride-sharing platform like Uber/Lyft where riders can request rides and get matched with nearby drivers in real-time.

## Functional Requirements
- Riders can request a ride from current location to destination
- System matches riders with nearby available drivers
- Riders and drivers see each other's real-time location
- Riders can see ETA and fare estimate before booking
- Payment is processed automatically after ride completion

## Non-Functional Requirements
- Low latency for matching (<5 seconds to find driver)
- Real-time location updates (<1 second lag)
- High availability (rides must complete even during partial outages)
- Scale to millions of concurrent rides globally

## Relevant Patterns

**Geospatial Indexing**: Store driver locations in geohash-based index for efficient nearby queries.

**Location Streaming**: Drivers continuously send location updates via WebSocket/long-poll.

**State Machine**: Ride goes through states (requested -> matched -> en route -> in progress -> completed).

## Common Deep Dives

**How do you efficiently find nearby drivers for a rider?**
- Use geohash to partition geographic space into cells
- Drivers report location to their current geohash cell
- On ride request, query current cell + neighboring cells
- Use Redis sorted sets: key=geohash, score=timestamp, value=driver_id
- Filter for available drivers, sort by distance/ETA
- Expand search radius if not enough drivers found

**How do you handle driver location updates at scale?**
- Drivers send location every 3-5 seconds
- WebSocket connection to regional gateway servers
- Batch location updates (3-5 updates per write)
- Update geohash index only on cell boundary crossing
- Store full location history to separate time-series DB for analytics
- Use regional sharding (drivers only need local matching)

**How do you implement surge pricing?**
- Track supply (available drivers) and demand (ride requests) per region
- Use smaller geohash cells for granular surge zones
- Calculate surge multiplier based on supply/demand ratio
- Update surge map every 1-2 minutes
- Show surge to rider before confirming ride
- Anti-gaming: ignore driver location spoofing, detect coordinated driver offline

**How do you ensure ride completion during failures?**
- Ride state stored durably (database, not just memory)
- Driver/rider apps cache current ride state locally
- On reconnect, sync state from server
- Payment can be processed asynchronously (queue-based)
- Multiple regional failovers for critical path

## Relevant Technologies
- **Redis**: Geospatial index (GEOADD/GEORADIUS), driver availability, surge data
- **Kafka**: Location event streaming, ride state changes, analytics
- **PostgreSQL/MySQL**: Ride records, user accounts, payment records
- **Google Maps API / OSRM**: ETA calculation, routing
- **WebSocket**: Real-time location updates between driver/rider

## Red Flags to Avoid

**Naive distance calculation**: "Query all drivers and calculate distance" doesn't scale.

**No state machine for rides**: Ride state must be explicit and durable for reliability.

**Ignoring location update scale**: Millions of drivers sending updates every few seconds is massive write load.

## Similar Problems to Practice
- Lyft
- DoorDash/Instacart (delivery matching)
- Yelp (nearby search)
- Google Maps (ETA, routing)
