package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) UpdateTrack(c *gin.Context, req *UpdateTrackRequest) *UpdateTrackResponse {
	track := &model.Track{
		TrackField: req.TrackField,
	}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := track.Update(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &UpdateTrackResponse{}
}

type UpdateTrackRequest struct {
	Id uint `json:"id"`
	model.TrackField
}

type UpdateTrackResponse struct {
}