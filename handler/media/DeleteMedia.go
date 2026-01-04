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
	deleted, err := DeleteObjects(c, req.Ids)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
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

func DeleteObjects(c *gin.Context, mediaUUIDs []uuid.UUID) ([]uuid.UUID, error) {
	deleted := []uuid.UUID{}
	client, err := service.InitLocalStorageService()
	if err != nil {
		return nil, err
	}
	for _, id := range mediaUUIDs {
		m := &model.Media{}
		err := m.Get(config.ContextDB(c), model.WhereMap{
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
	return deleted, nil
}
