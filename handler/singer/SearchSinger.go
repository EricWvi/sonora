package singer

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) SearchSinger(c *gin.Context, req *SearchSingerRequest) *SearchSingerResponse {
	if req.Query == "" {
		handler.Errorf(c, "query parameter is required")
		return nil
	}

	var where model.WhereExpr
	where.ILIKE(model.Singer_Name, "%"+req.Query+"%")

	singers, err := model.ListAllSingers(config.ContextDB(c), where)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &SearchSingerResponse{
		Singers: singers,
	}
}

type SearchSingerRequest struct {
	Query string `form:"query"`
}

type SearchSingerResponse struct {
	Singers []model.SingerView `json:"singers"`
}