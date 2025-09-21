package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListAlbumTracks(c *gin.Context, req *ListAlbumTracksRequest) *ListAlbumTracksResponse {
	if req.AlbumID == 0 {
		handler.Errorf(c, "album ID must be greater than 0")
		return nil
	}

	m := model.WhereMap{}
	m.Eq(model.Track_Album, req.AlbumID)

	tracks, err := model.ListAllTracks(config.ContextDB(c), m)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListAlbumTracksResponse{
		Tracks: tracks,
	}
}

type ListAlbumTracksRequest struct {
	AlbumID uint `form:"albumId" binding:"required"`
}

type ListAlbumTracksResponse struct {
	Tracks []model.TrackView `json:"tracks"`
}