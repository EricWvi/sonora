package sync

import (
	"slices"
	"testing"
	"time"

	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/tests"
	"github.com/EricWvi/sonora/tests/api"
	"github.com/stretchr/testify/assert"
)

func TestIncrementalSync_InsertAndUpdate(t *testing.T) {
	// Setup test router
	router := tests.TestRouter()

	// Step 1: Get initial sync timestamp
	initialSync := api.GetFullSync(t, router)
	syncTimestamp := initialSync.Timestamp

	// Wait a moment to ensure timestamp difference
	time.Sleep(100 * time.Millisecond)

	// Step 2: Create a new album (INSERT)
	albumCoverUUID := api.CreateTestMedia(t, router)
	albumID := api.CreateTestAlbum(t, router, model.AlbumField{
		Name:  "Incremental Test Album",
		Cover: albumCoverUUID.String(),
		Year:  2024,
	})
	defer api.DeleteTestAlbum(t, router, albumID)

	// Step 3: Update the album (UPDATE)
	time.Sleep(100 * time.Millisecond)
	api.UpdateTestAlbum(t, router, albumID, model.AlbumField{
		Name: "Updated Incremental Test Album",
	})

	// Step 4: Get incremental updates
	updatesResp := api.GetUpdates(t, router, syncTimestamp)

	// Step 5: Verify response
	assert.NotNil(t, updatesResp)
	assert.NotZero(t, updatesResp.Timestamp)
	assert.NotEmpty(t, updatesResp.Entries)

	// Step 6: Verify album appears in stale (not deleted)
	// INSERT -> UPDATE should appear as stale
	var foundAlbumUpdate bool
	for _, entry := range updatesResp.Entries {
		if entry.TableName == model.Album_Table {
			if slices.Contains(entry.Stale, albumID) {
				foundAlbumUpdate = true
			}
			// Should not be in deleted list
			for _, deletedID := range entry.Deleted {
				assert.NotEqual(t, albumID, deletedID, "album should not be in deleted list")
			}
		}
	}
	assert.True(t, foundAlbumUpdate, "updated album should appear in stale entries")
}

func TestIncrementalSync_InsertUpdateAndDelete(t *testing.T) {
	// Setup test router
	router := tests.TestRouter()

	// Step 1: Get initial sync timestamp
	initialSync := api.GetFullSync(t, router)
	syncTimestamp := initialSync.Timestamp

	// Wait a moment to ensure timestamp difference
	time.Sleep(100 * time.Millisecond)

	// Step 2: Create a new album (INSERT)
	albumCoverUUID := api.CreateTestMedia(t, router)
	albumID := api.CreateTestAlbum(t, router, model.AlbumField{
		Name:  "To Be Deleted Album",
		Cover: albumCoverUUID.String(),
		Year:  2024,
	})
	defer api.DeleteTestAlbum(t, router, albumID)

	// Step 3: Update the album (UPDATE)
	time.Sleep(100 * time.Millisecond)
	api.UpdateTestAlbum(t, router, albumID, model.AlbumField{
		Name: "Updated To Be Deleted Album",
	})

	// Step 4: Delete the album (DELETE)
	time.Sleep(100 * time.Millisecond)
	api.DeleteTestAlbum(t, router, albumID)

	// Step 5: Get incremental updates
	updatesResp := api.GetUpdates(t, router, syncTimestamp)

	// Step 6: Verify response
	assert.NotNil(t, updatesResp)
	assert.NotZero(t, updatesResp.Timestamp)
	assert.NotEmpty(t, updatesResp.Entries)

	// Step 7: Verify album appears in deleted (not stale)
	// INSERT -> UPDATE -> DELETE should appear as deleted
	var foundAlbumDelete bool
	for _, entry := range updatesResp.Entries {
		if entry.TableName == model.Album_Table {
			if slices.Contains(entry.Deleted, albumID) {
				foundAlbumDelete = true
			}
			// Should not be in stale list
			for _, staleID := range entry.Stale {
				assert.NotEqual(t, albumID, staleID, "album should not be in stale list")
			}
		}
	}
	assert.True(t, foundAlbumDelete, "deleted album should appear in deleted entries")
}

func TestIncrementalSync_MultipleRecords(t *testing.T) {
	// Setup test router
	router := tests.TestRouter()

	// Step 1: Get initial sync timestamp
	initialSync := api.GetFullSync(t, router)
	syncTimestamp := initialSync.Timestamp

	// Wait a moment to ensure timestamp difference
	time.Sleep(100 * time.Millisecond)

	// Step 2: Create multiple albums and tracks
	albumCoverUUID1 := api.CreateTestMedia(t, router)
	albumID1 := api.CreateTestAlbum(t, router, model.AlbumField{
		Name:  "Multi Test Album 1",
		Cover: albumCoverUUID1.String(),
		Year:  2024,
	})
	defer api.DeleteTestAlbum(t, router, albumID1)

	albumCoverUUID2 := api.CreateTestMedia(t, router)
	albumID2 := api.CreateTestAlbum(t, router, model.AlbumField{
		Name:  "Multi Test Album 2",
		Cover: albumCoverUUID2.String(),
		Year:  2024,
	})
	defer api.DeleteTestAlbum(t, router, albumID2)

	trackAudioUUID := api.CreateTestMedia(t, router)
	trackID := api.CreateTestTrack(t, router, model.TrackField{
		Name:     "Multi Test Track",
		Singer:   "Test Singer",
		Album:    int(albumID1),
		Cover:    albumCoverUUID1.String(),
		URL:      trackAudioUUID.String(),
		Duration: 180,
		Year:     2024,
		Genre:    "Test",
	})
	defer api.DeleteTestTrack(t, router, trackID)

	// Step 3: Update one album, delete another
	time.Sleep(100 * time.Millisecond)
	api.UpdateTestAlbum(t, router, albumID1, model.AlbumField{
		Name: "Updated Multi Test Album 1",
	})

	time.Sleep(100 * time.Millisecond)
	api.DeleteTestAlbum(t, router, albumID2)

	// Step 4: Get incremental updates
	updatesResp := api.GetUpdates(t, router, syncTimestamp)

	// Step 5: Verify response
	assert.NotNil(t, updatesResp)
	assert.NotZero(t, updatesResp.Timestamp)
	assert.NotEmpty(t, updatesResp.Entries)

	// Step 6: Verify changes
	var albumStaleIDs []uint
	var albumDeletedIDs []uint
	var trackStaleIDs []uint

	for _, entry := range updatesResp.Entries {
		if entry.TableName == model.Album_Table {
			albumStaleIDs = entry.Stale
			albumDeletedIDs = entry.Deleted
		}
		if entry.TableName == model.Track_Table {
			trackStaleIDs = entry.Stale
		}
	}

	// Album 1 should be in stale (created then updated)
	assert.Contains(t, albumStaleIDs, albumID1, "album1 should be in stale list")

	// Album 2 should be in deleted (created then deleted)
	assert.Contains(t, albumDeletedIDs, albumID2, "album2 should be in deleted list")

	// Track should be in stale (created)
	assert.Contains(t, trackStaleIDs, trackID, "track should be in stale list")
}

func TestIncrementalSync_NoChanges(t *testing.T) {
	// Setup test router
	router := tests.TestRouter()

	// Step 1: Get initial sync timestamp
	initialSync := api.GetFullSync(t, router)
	syncTimestamp := initialSync.Timestamp

	// Step 2: Immediately get updates without making any changes
	updatesResp := api.GetUpdates(t, router, syncTimestamp)

	// Step 3: Verify response is empty or has no entries
	assert.NotNil(t, updatesResp)
	assert.NotZero(t, updatesResp.Timestamp)
	// Entries should be empty or contain no changes
	if len(updatesResp.Entries) > 0 {
		for _, entry := range updatesResp.Entries {
			assert.Empty(t, entry.Stale, "should have no stale entries")
			assert.Empty(t, entry.Deleted, "should have no deleted entries")
		}
	}
}
