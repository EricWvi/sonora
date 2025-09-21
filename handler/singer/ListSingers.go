package singer

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListSingers(c *gin.Context, req *ListSingersRequest) *ListSingersResponse {
	m := model.WhereExpr{}

	singers, hasMore, err := model.ListSingers(config.ContextDB(c), m, req.Page)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListSingersResponse{
		Singers: singers,
		HasMore: hasMore,
	}
}

type ListSingersRequest struct {
	Page      uint   `form:"page"`
	Condition string `form:"condition"`
}

type QueryCondition struct {
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type ListSingersResponse struct {
	Singers []model.SingerView `json:"singers"`
	HasMore bool               `json:"hasMore"`
}
