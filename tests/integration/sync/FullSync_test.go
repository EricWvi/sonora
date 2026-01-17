package sync

import (
	"testing"

	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/tests"
	"github.com/EricWvi/sonora/tests/api"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	config.Init()
	m.Run()
}

func TestFullSync(t *testing.T) {
	// Setup test router
	router := tests.TestRouter()

	// Step 1: Create test data
	albumCoverUUID := api.CreateTestMedia(t, router)
	track1AudioUUID := api.CreateTestMedia(t, router)

	albumID := api.CreateTestAlbum(t, router, model.AlbumField{
		Name:  "Full Sync Test Album",
		Cover: albumCoverUUID.String(),
		Year:  2024,
	})
	defer api.DeleteTestAlbum(t, router, albumID)

	trackID := api.CreateTestTrack(t, router, model.TrackField{
		Name:     "Full Sync Test Track",
		Singer:   "Test Singer",
		Album:    int(albumID),
		Cover:    albumCoverUUID.String(),
		URL:      track1AudioUUID.String(),
		Duration: 180,
		Year:     2024,
		Genre:    "Test",
	})
	defer api.DeleteTestTrack(t, router, trackID)

	// Step 2: Call GetFullSync
	fullSyncResp := api.GetFullSync(t, router)

	// Step 3: Verify response structure
	assert.NotNil(t, fullSyncResp)
	assert.NotNil(t, fullSyncResp.Albums)
	assert.NotNil(t, fullSyncResp.Tracks)
	assert.NotNil(t, fullSyncResp.Singers)
	assert.NotNil(t, fullSyncResp.Lyrics)
	assert.NotZero(t, fullSyncResp.Timestamp)
	// Step 4: Verify our test data is included
	var foundAlbum bool
	for _, album := range fullSyncResp.Albums {
		if album.ID == albumID {
			foundAlbum = true
			assert.Equal(t, "Full Sync Test Album", album.Name)
			assert.Equal(t, albumCoverUUID.String(), album.Cover)
			assert.Equal(t, int16(2024), album.Year)
			break
		}
	}
	assert.True(t, foundAlbum, "created album should be in full sync response")

	var foundTrack bool
	for _, track := range fullSyncResp.Tracks {
		if track.ID == trackID {
			foundTrack = true
			assert.Equal(t, "Full Sync Test Track", track.Name)
			assert.Equal(t, int(albumID), track.Album)
			break
		}
	}
	assert.True(t, foundTrack, "created track should be in full sync response")
}
