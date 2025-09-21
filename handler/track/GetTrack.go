package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) GetTrack(c *gin.Context, req *GetTrackRequest) *GetTrackResponse {
	track := &model.Track{}
	m := model.WhereMap{model.Id: req.ID}

	if err := track.Get(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &GetTrackResponse{
		Track: model.TrackView{
			ID:         track.ID,
			TrackField: track.TrackField,
		},
	}
}

type GetTrackRequest struct {
	ID uint `form:"id" binding:"required"`
}

type GetTrackResponse struct {
	Track model.TrackView `json:"track"`
}