package singer

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) GetSinger(c *gin.Context, req *GetSingerRequest) *GetSingerResponse {
	singer := &model.Singer{}
	m := model.WhereMap{model.Id: req.ID}

	if err := singer.Get(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &GetSingerResponse{
		Singer: model.SingerView{
			ID:          singer.ID,
			SingerField: singer.SingerField,
		},
	}
}

type GetSingerRequest struct {
	ID uint `form:"id" binding:"required"`
}

type GetSingerResponse struct {
	Singer model.SingerView `json:"singer"`
}