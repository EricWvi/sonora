package singer

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) UpdateSinger(c *gin.Context, req *UpdateSingerRequest) *UpdateSingerResponse {
	singer := &model.Singer{
		SingerField: req.SingerField,
	}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := singer.Update(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &UpdateSingerResponse{}
}

type UpdateSingerRequest struct {
	Id uint `json:"id"`
	model.SingerField
}

type UpdateSingerResponse struct {
}
