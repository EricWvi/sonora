package singer

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) CreateSinger(c *gin.Context, req *CreateSingerRequest) *CreateSingerResponse {
	singer := &model.Singer{}
	singer.SingerField = req.SingerField

	if err := singer.Create(config.ContextDB(c)); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &CreateSingerResponse{
		Id: singer.ID,
	}
}

type CreateSingerRequest struct {
	model.SingerField
}

type CreateSingerResponse struct {
	Id uint `json:"id"`
}
