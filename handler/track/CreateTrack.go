package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) CreateTrack(c *gin.Context, req *CreateTrackRequest) *CreateTrackResponse {
	track := &model.Track{}
	track.TrackField = req.TrackField

	if err := track.Create(config.ContextDB(c)); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &CreateTrackResponse{
		Id: track.ID,
	}
}

type CreateTrackRequest struct {
	model.TrackField
}

type CreateTrackResponse struct {
	Id uint `json:"id"`
}