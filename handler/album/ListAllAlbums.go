package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListAllAlbums(c *gin.Context, req *ListAllAlbumsRequest) *ListAllAlbumsResponse {
	m := model.WhereExpr{}

	albums, err := model.ListAllAlbums(config.ContextDB(c), m)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListAllAlbumsResponse{
		Albums: albums,
	}
}

type ListAllAlbumsRequest struct {
}

type ListAllAlbumsResponse struct {
	Albums []model.AlbumView `json:"albums"`
}
