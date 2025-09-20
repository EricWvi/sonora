package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListTracks(c *gin.Context, req *ListTracksRequest) *ListTracksResponse {
	m := model.WhereMap{}

	tracks, err := model.ListTracks(config.ContextDB(c), m)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListTracksResponse{
		Tracks: tracks,
	}
}

type ListTracksRequest struct {
}

type ListTracksResponse struct {
	Tracks []model.Track `json:"tracks"`
}