# Server Development Log

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
