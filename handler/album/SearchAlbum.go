package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) SearchAlbum(c *gin.Context, req *SearchAlbumRequest) *SearchAlbumResponse {
	if req.Query == "" {
		handler.Errorf(c, "query parameter is required")
		return nil
	}

	var where model.WhereExpr
	where.ILIKE(model.Album_Name, "%"+req.Query+"%")

	albums, err := model.ListAllAlbums(config.ContextDB(c), where)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &SearchAlbumResponse{
		Albums: albums,
	}
}

type SearchAlbumRequest struct {
	Query string `form:"query"`
}

type SearchAlbumResponse struct {
	Albums []model.AlbumView `json:"albums"`
}