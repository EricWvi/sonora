package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListSingles(c *gin.Context, req *ListSinglesRequest) *ListSinglesResponse {
	m := model.WhereMap{}
	m.Eq(model.Track_Album, 0)

	tracks, err := model.ListAllTracks(config.ContextDB(c), m)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListSinglesResponse{
		Tracks: tracks,
	}
}

type ListSinglesRequest struct {
}

type ListSinglesResponse struct {
	Tracks []model.TrackView `json:"tracks"`
}