package singer

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListSingers(c *gin.Context, req *ListSingersRequest) *ListSingersResponse {
	m := model.WhereMap{}

	singers, err := model.ListSingers(config.ContextDB(c), m)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListSingersResponse{
		Singers: singers,
	}
}

type ListSingersRequest struct {
}

type ListSingersResponse struct {
	Singers []model.Singer `json:"singers"`
}
