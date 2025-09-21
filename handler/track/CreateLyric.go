package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) CreateLyric(c *gin.Context, req *CreateLyricRequest) *CreateLyricResponse {
	lyric := &model.Lyric{}
	lyric.LyricField = req.LyricField

	if err := lyric.Create(config.ContextDB(c)); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &CreateLyricResponse{
		Id: lyric.ID,
	}
}

type CreateLyricRequest struct {
	model.LyricField
}

type CreateLyricResponse struct {
	Id uint `json:"id"`
}