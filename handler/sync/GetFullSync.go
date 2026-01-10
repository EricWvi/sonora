package sync

import (
	"time"

	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) GetFullSync(c *gin.Context, req *GetFullSyncRequest) *GetFullSyncResponse {
	db := config.ContextDB(c)
	m := model.WhereExpr{}

	// Fetch all albums
	albums, err := model.ListAllAlbums(db, m)
	if err != nil {
		handler.Errorf(c, "failed to fetch albums: %s", err.Error())
		return nil
	}

	// Fetch all lyrics
	lyrics, err := model.ListAllLyrics(db, m)
	if err != nil {
		handler.Errorf(c, "failed to fetch lyrics: %s", err.Error())
		return nil
	}

	// Fetch all singers
	singers, err := model.ListAllSingers(db, m)
	if err != nil {
		handler.Errorf(c, "failed to fetch singers: %s", err.Error())
		return nil
	}

	// Fetch all tracks
	tracks, err := model.ListAllTracks(db, m)
	if err != nil {
		handler.Errorf(c, "failed to fetch tracks: %s", err.Error())
		return nil
	}

	return &GetFullSyncResponse{
		Data: model.FullSyncResponse{
			Albums:    albums,
			Lyrics:    lyrics,
			Singers:   singers,
			Tracks:    tracks,
			Timestamp: time.Now().UnixMilli(),
		},
	}
}

type GetFullSyncRequest struct {
}

type GetFullSyncResponse struct {
	Data model.FullSyncResponse `json:"data"`
}
