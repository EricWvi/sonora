package singer

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListAllSingers(c *gin.Context, req *ListAllSingersRequest) *ListAllSingersResponse {
	m := model.WhereMap{}

	singers, err := model.ListAllSingers(config.ContextDB(c), m)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListAllSingersResponse{
		Singers: singers,
	}
}

type ListAllSingersRequest struct {
}

type ListAllSingersResponse struct {
	Singers []model.SingerView `json:"singers"`
}
