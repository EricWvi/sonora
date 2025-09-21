package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListAlbums(c *gin.Context, req *ListAlbumsRequest) *ListAlbumsResponse {
	m := model.WhereExpr{}

	albums, hasMore, err := model.ListAlbums(config.ContextDB(c), m, req.Page)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListAlbumsResponse{
		Albums:  albums,
		HasMore: hasMore,
	}
}

type ListAlbumsRequest struct {
	Page      uint   `form:"page"`
	Condition string `form:"condition"`
}

type QueryCondition struct {
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type ListAlbumsResponse struct {
	Albums  []model.AlbumView `json:"albums"`
	HasMore bool              `json:"hasMore"`
}