# Design Dropbox Guide

Design a file storage and synchronization service like Dropbox/Google Drive where users can upload files and have them sync across all their devices.

## Functional Requirements
- Users can upload, download, and delete files/folders
- Files sync automatically across all user's devices
- Users can share files/folders with others (view or edit)
- Users can view file version history and restore previous versions

## Non-Functional Requirements
- Reliable sync (no data loss, no corruption)
- Efficient bandwidth usage (don't re-upload unchanged data)
- Low sync latency (changes appear on other devices quickly)
- Scale to billions of files, petabytes of storage

## Relevant Patterns

**Chunking**: Split large files into fixed-size chunks (4-8MB). Enables deduplication and resumable uploads.

**Content-Addressable Storage**: Store chunks by their hash. Same content = same hash = store once.

**Metadata vs Content Separation**: Metadata service tracks file structure; separate blob storage for content.

## Common Deep Dives

**How do you handle file uploads efficiently?**
- Split file into fixed-size chunks (4MB typical)
- Calculate hash (SHA-256) of each chunk
- Check which chunks already exist in storage (dedup check)
- Only upload chunks that don't exist
- After all chunks uploaded, commit metadata (file -> [chunk hashes])
- For small files, may inline content in metadata

**How do you detect file changes and sync efficiently?**
- Client watches filesystem for changes (inotify/FSEvents)
- On change, re-chunk the file
- Compare new chunk hashes with stored hashes
- Only upload changed chunks (delta sync)
- Notify sync server of new file version
- Other clients receive notification and pull changes

**How do you handle concurrent edits to the same file?**
- Each edit creates a new version (append-only)
- Conflict detection: check if base version matches server's latest
- If conflict: save both versions, let user resolve
- For collaborative editing: use OT/CRDT (separate problem from storage)
- Last-write-wins for simple cases (configurable)

**How do you ensure durability and availability of stored files?**
- Store each chunk with 3x replication across data centers
- Or use erasure coding (more space-efficient for cold data)
- Metadata stored in replicated database (MySQL/PostgreSQL)
- Regular integrity checks (scrubbing) to detect bit rot
- Backup metadata to separate region

## Relevant Technologies
- **S3/Blob Storage**: Content-addressable chunk storage
- **MySQL/PostgreSQL**: Metadata (users, files, folder structure, versions)
- **Redis**: Caching file metadata, chunk existence checks
- **Kafka**: Sync notifications, event streaming
- **WebSocket/Long-poll**: Real-time sync notifications to clients

## Red Flags to Avoid

**No chunking**: "Upload whole file on every change" wastes bandwidth and doesn't scale.

**No deduplication**: Same file uploaded by different users stored multiple times.

**No versioning**: Can't recover from accidental deletes or overwrites.

**Conflict = overwrite**: Silently losing user data on concurrent edits.

## Similar Problems to Practice
- Google Drive
- OneDrive
- iCloud Drive
- Box
