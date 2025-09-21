package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) UpdateLyric(c *gin.Context, req *UpdateLyricRequest) *UpdateLyricResponse {
	lyric := &model.Lyric{
		LyricField: req.LyricField,
	}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := lyric.Update(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &UpdateLyricResponse{}
}

type UpdateLyricRequest struct {
	Id uint `json:"id"`
	model.LyricField
}

type UpdateLyricResponse struct {
}