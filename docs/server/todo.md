## TODO: delete media when deleting track and album

Status: ✅ **COMPLETED** (2026-01-04)

## Goal
- also delete media's record and local storage when deleting a track
- when deleting album, delete track records, media records, and local storage.
- test the modified logic. (In test, always create a new record then delete it) For media upload, just put an array of bits in request body.

### Plan

- [x] Update `handler/track/DeleteTrack.go`
- [x] Update `handler/album/DeleteAlbum.go`
- [x] Create `tests/integration/track/DeleteTrack_test.go`
- [x] Create `tests/integration/album/DeleteAlbum_test.go`
- [x] Create `tests/api/album.go`
- [x] Create `tests/api/media.go`
- [x] Create `tests/api/track.go`

### Implementation Summary
All tasks completed successfully. Cascade deletion now properly removes:
- Track/Album database records
- Associated media database records
- Physical files from local storage

See `docs/server/dev-log.md` for detailed implementation notes.
