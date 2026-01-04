package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/handler/media"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (b Base) DeleteTrack(c *gin.Context, req *DeleteTrackRequest) *DeleteTrackResponse {
	db := config.ContextDB(c)

	// First, get the track to retrieve media UUIDs
	track := &model.Track{}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := track.Get(db, m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	// Collect media UUIDs to delete (cover and url)
	mediaUUIDs := []uuid.UUID{}
	// only delete cover if it's not album cover
	if track.Album == 0 && track.Cover != "" {
		if coverUUID, err := uuid.Parse(track.Cover); err == nil {
			mediaUUIDs = append(mediaUUIDs, coverUUID)
		}
	}
	if track.URL != "" {
		if urlUUID, err := uuid.Parse(track.URL); err == nil {
			mediaUUIDs = append(mediaUUIDs, urlUUID)
		}
	}

	// Delete the track record
	if err := track.Delete(db, m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	go func() {
		media.DeleteObjects(c, mediaUUIDs)
	}()

	return &DeleteTrackResponse{}
}

type DeleteTrackRequest struct {
	Id uint `json:"id"`
}

type DeleteTrackResponse struct {
}
