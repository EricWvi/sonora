package track

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

func TestDeleteTrack(t *testing.T) {
	config.Init()

	// Setup test router
	router := tests.TestRouter()

	// Step 1: Upload two media files (cover and audio)
	coverUUID := api.CreateTestMedia(t, router)
	audioUUID := api.CreateTestMedia(t, router)

	// Step 2: Create a track with the media
	trackID := api.CreateTestTrack(t, router, model.TrackField{
		Name:     "Test Track",
		Singer:   "Test Singer",
		Album:    0,
		Cover:    coverUUID.String(),
		URL:      audioUUID.String(),
		Duration: 180,
		Year:     2024,
		Genre:    "Test",
	})
	defer api.DeleteTestTrack(t, router, trackID)

	// Step 3: Verify media files exist
	storage, err := service.InitLocalStorageService()
	assert.NoError(t, err)

	db := config.DB()
	coverMedia := &model.Media{}
	err = coverMedia.Get(db, model.WhereMap{model.Media_Link: coverUUID})
	assert.NoError(t, err)
	assert.True(t, storage.FileExists(coverMedia.Key), "cover file should exist before deletion")

	audioMedia := &model.Media{}
	err = audioMedia.Get(db, model.WhereMap{model.Media_Link: audioUUID})
	assert.NoError(t, err)
	assert.True(t, storage.FileExists(audioMedia.Key), "audio file should exist before deletion")

	// Step 4: Delete the track
	api.DeleteTestTrack(t, router, trackID)
	time.Sleep(1 * time.Second) // wait for async deletion

	// Step 5: Verify track is deleted
	track := &model.Track{}
	err = track.Get(db, model.WhereMap{model.Id: trackID})
	assert.Error(t, err, "track should not exist after deletion")

	// Step 6: Verify media records are deleted
	coverMediaAfter := &model.Media{}
	err = coverMediaAfter.Get(db, model.WhereMap{model.Media_Link: coverUUID})
	assert.Error(t, err, "cover media record should be deleted")

	audioMediaAfter := &model.Media{}
	err = audioMediaAfter.Get(db, model.WhereMap{model.Media_Link: audioUUID})
	assert.Error(t, err, "audio media record should be deleted")

	// Step 7: Verify files are deleted from storage
	assert.False(t, storage.FileExists(coverMedia.Key), "cover file should be deleted from storage")
	assert.False(t, storage.FileExists(audioMedia.Key), "audio file should be deleted from storage")
}
