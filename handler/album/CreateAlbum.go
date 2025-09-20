package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) CreateAlbum(c *gin.Context, req *CreateAlbumRequest) *CreateAlbumResponse {
	album := &model.Album{}
	album.AlbumField = req.AlbumField

	if err := album.Create(config.ContextDB(c)); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &CreateAlbumResponse{
		Id: album.ID,
	}
}

type CreateAlbumRequest struct {
	model.AlbumField
}

type CreateAlbumResponse struct {
	Id uint `json:"id"`
}