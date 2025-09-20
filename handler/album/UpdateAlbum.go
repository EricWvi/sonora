package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) UpdateAlbum(c *gin.Context, req *UpdateAlbumRequest) *UpdateAlbumResponse {
	album := &model.Album{
		AlbumField: req.AlbumField,
	}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := album.Update(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &UpdateAlbumResponse{}
}

type UpdateAlbumRequest struct {
	Id uint `json:"id"`
	model.AlbumField
}

type UpdateAlbumResponse struct {
}