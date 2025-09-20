package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) DeleteAlbum(c *gin.Context, req *DeleteAlbumRequest) *DeleteAlbumResponse {
	album := &model.Album{}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := album.Delete(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &DeleteAlbumResponse{}
}

type DeleteAlbumRequest struct {
	Id uint `json:"id"`
}

type DeleteAlbumResponse struct {
}