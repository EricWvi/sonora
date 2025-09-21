package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) GetLyric(c *gin.Context, req *GetLyricRequest) *GetLyricResponse {
	lyric := &model.Lyric{}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := lyric.Get(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &GetLyricResponse{
		Lyric: lyric.Content,
	}
}

type GetLyricRequest struct {
	Id uint `form:"id"`
}

type GetLyricResponse struct {
	Lyric string `json:"lyric"`
}