package media

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/log"
	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (b Base) DeleteMedia(c *gin.Context, req *DeleteMediaRequest) *DeleteMediaResponse {
	deleted := []uuid.UUID{}
	client, err := service.InitLocalStorageService()
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}
	for _, id := range req.Ids {
		m := &model.Media{}
		err := m.Get(config.ContextDB(c), gin.H{
			model.Media_Link: id,
		})
		if err != nil {
			continue
		}
		err = m.Delete(config.ContextDB(c), nil)
		if err != nil {
			log.Errorf(c, "DeleteMedia %s failed: %s", id, err)
			continue
		}
		err = client.DeleteObject(c, m.Key)
		if err != nil {
			log.Errorf(c, "DeleteObject %s failed: %s", m.Key, err)
			continue
		}
		log.Infof(c, "Object %s deleted successfully", m.Key)
		deleted = append(deleted, id)
	}

	return &DeleteMediaResponse{
		Ids: deleted,
	}
}

type DeleteMediaRequest struct {
	Ids []uuid.UUID `json:"ids"`
}

type DeleteMediaResponse struct {
	Ids []uuid.UUID `json:"ids"`
}
