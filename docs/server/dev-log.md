# Server Development Log

## 2026-01-11: Sync API Integration Tests - Part 3

### Summary
Updated timestamp precision from Unix seconds to Unix milliseconds to fix test timing issues.

### Changes Made

#### 1. Migration - Millisecond Timestamps
- **File**: `migration/migrations.go` (UPDATED - v0.10.0)
  - Changed `changed_at` default from `EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)` to `(EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000)::BIGINT`
  - Now stores timestamps in milliseconds instead of seconds
  - Fixes issue where fast operations in same second weren't captured

#### 2. Handler Updates
- **File**: `handler/sync/GetFullSync.go` (UPDATED)
  - Changed from `time.Now().Unix()` to `time.Now().UnixMilli()`
  - Returns millisecond timestamp for precise sync points

- **File**: `handler/sync/GetUpdates.go` (UPDATED)
  - Changed from `time.Now().Unix()` to `time.Now().UnixMilli()`
  - Consistent millisecond precision throughout sync API

### Test Results

✅ **All Tests Passing**:
- `TestFullSync` ✅
- `TestIncrementalSync_InsertAndUpdate` ✅
- `TestIncrementalSync_InsertUpdateAndDelete` ✅ (FIXED)
- `TestIncrementalSync_MultipleRecords` ✅
- `TestIncrementalSync_NoChanges` ✅ (FIXED)

### Issues Resolved

1. **Timestamp Precision**: Operations happening within the same second are now tracked separately
2. **Test Isolation**: Millisecond precision prevents changelog entries from bleeding between tests

### Technical Notes

The change from Unix seconds (10 digits) to Unix milliseconds (13 digits):
- Before: `1768099371` (second precision)
- After: `1768099371378` (millisecond precision)

This provides sub-second granularity needed for fast test execution and rapid client sync operations.

## 2026-01-11: Sync API Integration Tests - Part 2

### Summary
Updated migration system to support running individual migrations via CLI, fixed soft delete handling in changelog triggers, and added defer cleanup in tests.

### Changes Made

#### 1. Migration CLI Support
- **File**: `main.go` (UPDATED)
  - Added command-line parsing for `migrate` command
  - Supports `go run . migrate --version v0.10.0 --action up/down`
  - Migration runs and exits without starting server
  
- **File**: `migration/migration.go` (UPDATED)
  - Added `RunOne()` method to run specific migration version
  - Supports both up and down migrations
  - Validates migration exists and state before running

#### 2. Soft Delete Support in Changelog Triggers
- **File**: `migration/migrations.go` (UPDATED  - v0.10.0)
  - Updated `log_table_changes()` trigger function
  - Now detects GORM soft deletes (`deleted_at` changed from NULL to timestamp)
  - Logs soft deletes as DELETE operations instead of UPDATE
  - Maintains backward compatibility with hard deletes

#### 3. Test Cleanup with Defer
- **File**: `tests/api/album.go` (UPDATED)
  - `DeleteTestAlbum()` now checks if record exists before deletion
  - Prevents errors when defer tries to delete already-deleted records
  
- **File**: `tests/api/track.go` (UPDATED)
  - `DeleteTestTrack()` same improvements as album
  
- **Files**: All sync test files (UPDATED)
  - Added `defer` cleanup statements after creating test records
  - Ensures test isolation and cleanup even if test fails

### Test Results

✅ **Passing Tests**:
- `TestFullSync` - Full sync functionality with cleanup ✅
- `TestIncrementalSync_InsertAndUpdate` - INSERT→UPDATE correctly returns as stale ✅
- `TestIncrementalSync_MultipleRecords` - Multiple operations tracked correctly ✅

⚠️ **Known Issues**:
- Two tests have timestamp precision issues
- Operations happening in same second not captured by `>` comparison
- Need to either:
  - Change query to `>=` (but need to handle duplicate syncs)
  - Add sub-second precision to timestamps
  - Add sleep delays between operations (not ideal)

### Technical Details

The soft delete trigger now uses this logic:
```sql
IF (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL) THEN
    -- This is a soft delete
    INSERT INTO d_change_log VALUES (TG_TABLE_NAME, NEW.id, 'DELETE');
ELSE
    -- Regular update
    INSERT INTO d_change_log VALUES (TG_TABLE_NAME, NEW.id, 'UPDATE');
END IF;
```

### Migration Command Examples
```bash
# Run a specific migration up
go run . migrate --version v0.10.0 --action up

# Rollback a specific migration
go run . migrate --version v0.10.0 --action down
```

## 2026-01-11: Sync API Integration Tests

### Summary
Implemented comprehensive integration tests for the sync API endpoints to verify client-side caching functionality works correctly with both full and incremental sync.

### Changes Made

#### 1. Test API Helpers
- **File**: `tests/api/sync.go` (NEW)
  - `GetFullSync()` - Helper to call full sync endpoint and parse response
  - `GetUpdates()` - Helper to call incremental sync endpoint with timestamp parameter
  
- **File**: `tests/api/album.go` (UPDATED)
  - `UpdateTestAlbum()` - New helper for album updates in tests
  - Follows existing pattern with CreateTestAlbum and DeleteTestAlbum

#### 2. Integration Tests
- **File**: `tests/integration/sync/FullSync_test.go` (NEW)
  - Tests full sync returns complete dataset
  - Verifies created test data appears in sync response
  - Validates response structure (albums, tracks, singers, lyrics, timestamp)
  
- **File**: `tests/integration/sync/IncrementalSync_test.go` (NEW)
  - `TestIncrementalSync_InsertAndUpdate` - Verifies INSERT→UPDATE appears as stale ✅
  - `TestIncrementalSync_InsertUpdateAndDelete` - Tests INSERT→UPDATE→DELETE sequence
  - `TestIncrementalSync_MultipleRecords` - Tests multiple records and operations
  - `TestIncrementalSync_NoChanges` - Verifies empty response when no changes

### Test Results

✅ **Passing Tests**:
- `TestFullSync` - Full sync functionality works correctly
- `TestIncrementalSync_InsertAndUpdate` - INSERT→UPDATE correctly returns as stale

⚠️ **Failing Tests** (Need Investigation):
- Some tests failing due to soft delete interaction with changelog triggers
- GORM soft deletes update `deleted_at` timestamp rather than actual DELETE
- Changelog trigger may need adjustment to detect soft deletes as DELETE operations


## 2026-01-10: Client-Side Caching with Incremental Sync

### Summary
Implemented a comprehensive client-side caching system with incremental sync capabilities. The system uses PostgreSQL triggers to automatically track all changes to music data, enabling efficient client-side cache updates without polling or full data re-fetches.

### Changes Made

#### 1. Database Infrastructure (Migration v0.10.0)
- **File**: `migration/migrations.go`
- Created `d_change_log` table to track all data modifications
- Created PostgreSQL trigger function `log_table_changes()` that logs all operations
- Applied triggers to all tracked tables: `d_album`, `d_lyric`, `d_singer`, `d_track`
- All changes are automatically logged without application code intervention

#### 2. Model Layer
- **New File**: `model/changelog.go`

#### 3. Handler Layer
- **New Directory**: `handler/sync/`
- **File**: `handler/sync/base.go` - Sync handler base with dispatcher
- **File**: `handler/sync/GetFullSync.go` - Full sync endpoint:
  - Fetches all records from all tracked tables
  - Returns complete dataset with current unix timestamp
  - Used for initial cache population
- **File**: `handler/sync/GetUpdates.go` - Incremental sync endpoint:
  - Accepts `since` unix timestamp parameter
  - Fetches all change logs since the given time
  - Deduplicates changes (only latest operation per record)
  - Returns array of stale ids and deleted ids

#### 4. Router Integration
- **File**: `router.go`
- Registered sync handler: `GET /api/sync`
- Added import for `handler/sync` package

### Design Decisions

1. **Trigger-Based Change Tracking**:
   - Uses PostgreSQL triggers instead of application-level tracking
   - Guarantees all changes are logged, even from direct database access
   - Zero application overhead for change logging
   - Atomic - changes and logs are committed together

2. **Change Deduplication**:
   - Multiple operations on the same record only return the latest state
   - Example: INSERT → UPDATE → UPDATE returns only the final UPDATE
   - Example: INSERT → UPDATE → DELETE returns only the DELETE
   - Minimizes bandwidth and client processing

3. **Timestamp-Based Sync**:
   - Uses server-side timestamps from database
   - Client stores last sync timestamp and requests changes since then

4. **Full vs Incremental Sync**:
   - Full sync for initial cache population
   - Incremental sync for subsequent updates
   - Clear separation of concerns


## 2026-01-04: Cascade Delete Implementation for Media

### Summary
Implemented cascade deletion for tracks and albums to properly clean up associated media records and files from local storage. Refactored tests into a shared `tests/` directory with reusable API helpers.

### Changes Made

#### 1. Updated DeleteTrack and DeleteAlbum Handlers
- **File**: `handler/track/DeleteTrack.go`, `handler/album/DeleteAlbum.go`
- Smart cover handling: Only deletes track's cover if it's a standalone track (album == 0), avoiding deletion of shared album covers
- Asynchronous media cleanup using goroutine to avoid blocking the response

#### 2. Test Infrastructure Improvements
- **New Directory**: `tests/` - Centralized test utilities
  - `tests/tests.go` - Core test router setup with Action middleware
  - `tests/api/media.go` - Media creation helper
  - `tests/api/track.go` - Track creation/deletion helpers
  - `tests/api/album.go` - Album creation/deletion helpers

- **Test Files**: Moved to `tests/integration/` for better organization
  - `tests/integration/track/DeleteTrack_test.go` - Track deletion integration test
  - `tests/integration/album/DeleteAlbum_test.go` - Album deletion integration test

- **Reusable Test Helpers**:
  - `TestRouter()` - Pre-configured gin router with Action middleware
  - `ServeJSON()` - Generic JSON request helper
  - `CreateTestMedia()` - Multipart file upload helper
  - `CreateTestTrack()` - Track creation with model.TrackField
  - `DeleteTestTrack()` - Track deletion helper
  - `CreateTestAlbum()` - Album creation with model.AlbumField  
  - `DeleteTestAlbum()` - Album deletion helper

### Design Decisions

1. **Smart Cover Management**: 
   - Tracks with album (album != 0) share the album's cover - track deletion skips cover deletion
   - Standalone tracks (album == 0) have their own covers - track deletion removes the cover
   - Album deletion removes the album cover (shared by all tracks)

2. **Asynchronous Media Cleanup**: 
   - Media files are deleted asynchronously via goroutines after the main entity deletion
   - Provides faster API response times
   - Non-blocking: File system errors don't affect database operations

3. **Test Organization**:
   - Separated test utilities (`tests/`) from test cases (`tests/integration/`)
   - Reusable API helpers reduce boilerplate in test files

4. **Error Handling**: 
   - Media deletion failures are logged but don't block the primary operation
   - Individual media deletion failures don't stop batch processing
   - Track/album deletion succeeds even if media cleanup partially fails

### Testing
Integration tests in `tests/integration/` verify the complete deletion pipeline:
- **TestDeleteTrack**: Creates track with cover and audio, verifies complete cleanup
- **TestDeleteAlbum**: Creates album with 2 tracks, verifies cascade deletion of all entities

Both tests verify:
- Database records (tracks, albums, media) are deleted
- Physical files are removed from storage
- Proper handling of shared vs. standalone covers
