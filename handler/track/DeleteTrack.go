package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) DeleteTrack(c *gin.Context, req *DeleteTrackRequest) *DeleteTrackResponse {
	track := &model.Track{}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := track.Delete(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &DeleteTrackResponse{}
}

type DeleteTrackRequest struct {
	Id uint `json:"id"`
}

type DeleteTrackResponse struct {
}