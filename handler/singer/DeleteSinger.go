package singer

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) DeleteSinger(c *gin.Context, req *DeleteSingerRequest) *DeleteSingerResponse {
	singer := &model.Singer{}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := singer.Delete(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &DeleteSingerResponse{}
}

type DeleteSingerRequest struct {
	Id uint `json:"id"`
}

type DeleteSingerResponse struct {
}
