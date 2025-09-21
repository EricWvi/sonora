package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) GetAlbum(c *gin.Context, req *GetAlbumRequest) *GetAlbumResponse {
	album := &model.Album{}
	m := model.WhereMap{model.Id: req.ID}

	if err := album.Get(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &GetAlbumResponse{
		Album: model.AlbumView{
			ID:         album.ID,
			AlbumField: album.AlbumField,
		},
	}
}

type GetAlbumRequest struct {
	ID uint `form:"id" binding:"required"`
}

type GetAlbumResponse struct {
	Album model.AlbumView `json:"album"`
}