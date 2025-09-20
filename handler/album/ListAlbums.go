package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListAlbums(c *gin.Context, req *ListAlbumsRequest) *ListAlbumsResponse {
	m := model.WhereMap{}

	albums, err := model.ListAlbums(config.ContextDB(c), m)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListAlbumsResponse{
		Albums: albums,
	}
}

type ListAlbumsRequest struct {
}

type ListAlbumsResponse struct {
	Albums []model.Album `json:"albums"`
}