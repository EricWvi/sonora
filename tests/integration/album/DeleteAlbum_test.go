package album

import (
	"testing"
	"time"

	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/service"
	"github.com/EricWvi/sonora/tests"
	"github.com/EricWvi/sonora/tests/api"
	"github.com/stretchr/testify/assert"
)

func TestDeleteAlbum(t *testing.T) {
	config.Init()

	// Setup test router
	router := tests.TestRouter()

	// Step 1: Upload media files
	albumCoverUUID := api.CreateTestMedia(t, router)
	track1AudioUUID := api.CreateTestMedia(t, router)
	track2AudioUUID := api.CreateTestMedia(t, router)
	// Step 2: Create an album
	albumID := api.CreateTestAlbum(t, router, model.AlbumField{
		Name:  "Test Album",
		Cover: albumCoverUUID.String(),
		Year:  2024,
	})
	defer api.DeleteTestAlbum(t, router, albumID)

	// Step 3: Create tracks for the album
	track1ID := api.CreateTestTrack(t, router, model.TrackField{
		Name:     "Test Track",
		Singer:   "Test Singer",
		Album:    int(albumID),
		Cover:    albumCoverUUID.String(),
		URL:      track1AudioUUID.String(),
		Duration: 180,
		Year:     2024,
		Genre:    "Test",
	})
	track2ID := api.CreateTestTrack(t, router, model.TrackField{
		Name:     "Test Track 2",
		Singer:   "Test Singer",
		Album:    int(albumID),
		Cover:    albumCoverUUID.String(),
		URL:      track2AudioUUID.String(),
		Duration: 200,
		Year:     2024,
		Genre:    "Test",
	})

	// Step 4: Verify all media files exist
	storage, err := service.InitLocalStorageService()
	assert.NoError(t, err)

	db := config.DB()

	// Verify album cover
	albumCoverMedia := &model.Media{}
	err = albumCoverMedia.Get(db, model.WhereMap{model.Media_Link: albumCoverUUID})
	assert.NoError(t, err)
	assert.True(t, storage.FileExists(albumCoverMedia.Key), "album cover should exist before deletion")

	// Verify track 1 media
	track1AudioMedia := &model.Media{}
	err = track1AudioMedia.Get(db, model.WhereMap{model.Media_Link: track1AudioUUID})
	assert.NoError(t, err)
	assert.True(t, storage.FileExists(track1AudioMedia.Key), "track 1 audio should exist before deletion")

	// Verify track 2 media
	track2AudioMedia := &model.Media{}
	err = track2AudioMedia.Get(db, model.WhereMap{model.Media_Link: track2AudioUUID})
	assert.NoError(t, err)
	assert.True(t, storage.FileExists(track2AudioMedia.Key), "track 2 audio should exist before deletion")

	// Step 5: Delete the album
	api.DeleteTestAlbum(t, router, albumID)
	time.Sleep(1 * time.Second) // wait for async deletion

	// Step 6: Verify album is deleted
	album := &model.Album{}
	err = album.Get(db, model.WhereMap{model.Id: albumID})
	assert.Error(t, err, "album should not exist after deletion")

	// Step 7: Verify all tracks are deleted
	track1 := &model.Track{}
	err = track1.Get(db, model.WhereMap{model.Id: track1ID})
	assert.Error(t, err, "track 1 should be deleted")

	track2 := &model.Track{}
	err = track2.Get(db, model.WhereMap{model.Id: track2ID})
	assert.Error(t, err, "track 2 should be deleted")

	// Step 8: Verify all media records are deleted
	err = albumCoverMedia.Get(db, model.WhereMap{model.Media_Link: albumCoverUUID})
	assert.Error(t, err, "album cover media record should be deleted")

	err = track1AudioMedia.Get(db, model.WhereMap{model.Media_Link: track1AudioUUID})
	assert.Error(t, err, "track 1 audio media record should be deleted")

	err = track2AudioMedia.Get(db, model.WhereMap{model.Media_Link: track2AudioUUID})
	assert.Error(t, err, "track 2 audio media record should be deleted")

	// Step 9: Verify all files are deleted from storage
	assert.False(t, storage.FileExists(albumCoverMedia.Key), "album cover file should be deleted from storage")
	assert.False(t, storage.FileExists(track1AudioMedia.Key), "track 1 audio file should be deleted from storage")
	assert.False(t, storage.FileExists(track2AudioMedia.Key), "track 2 audio file should be deleted from storage")
}
