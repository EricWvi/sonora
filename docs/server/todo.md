## TODO: delete media when deleting track

Status: not implemented

## Goal
- also delete media's record and local storage when deleting a track
- when deleting album, delete track records, media records, and local storage.
- test the modified logic. (In test, always create a new record then delete it) For media upload, just put an array of bits in request body.

### Plan

- [ ] Update `model/track.go`
- [ ] Update `model/album.go`
- [ ] Update `config/db.go`: func for setting up test env
- [ ] Create `model/track_test.go`
- [ ] Create `model/album_test.go`


## Review

### Summary of Changes


### Files Changed

| File | Action |
|------|--------|
| | |
