package album

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/handler/media"
	"github.com/EricWvi/sonora/log"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (b Base) DeleteAlbum(c *gin.Context, req *DeleteAlbumRequest) *DeleteAlbumResponse {
	db := config.ContextDB(c)

	// First, get the album to retrieve its cover media UUID
	album := &model.Album{}
	m := model.WhereMap{}
	m.Eq(model.Id, req.Id)

	if err := album.Get(db, m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	// Get all tracks associated with this album
	whereExpr := model.WhereExpr{}
	whereExpr.Eq(model.Track_Album, req.Id)
	tracks, err := model.ListAllTracks(db, whereExpr)
	if err != nil {
		handler.Errorf(c, "failed to list album tracks: %s", err.Error())
		return nil
	}

	// Collect all media UUIDs to delete (from tracks and album cover)
	mediaUUIDs := []uuid.UUID{}

	// Add album cover
	if album.Cover != "" {
		if coverUUID, err := uuid.Parse(album.Cover); err == nil {
			mediaUUIDs = append(mediaUUIDs, coverUUID)
		}
	}

	// Add track media (url for each track)
	trackIDs := []uint{}
	for _, trackView := range tracks {
		trackIDs = append(trackIDs, trackView.ID)
		// since tracks of an album share same album cover, no need to add album cover again
		if trackView.URL != "" {
			if urlUUID, err := uuid.Parse(trackView.URL); err == nil {
				mediaUUIDs = append(mediaUUIDs, urlUUID)
			}
		}
	}

	// Delete all tracks belonging to this album
	for _, trackID := range trackIDs {
		track := &model.Track{}
		trackWhere := model.WhereMap{}
		trackWhere.Eq(model.Id, trackID)
		if err := track.Delete(db, trackWhere); err != nil {
			log.Errorf(c, "failed to delete track %d: %s", trackID, err)
		}
	}

	// Delete the album record
	if err := album.Delete(db, m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	go func() {
		media.DeleteObjects(c, mediaUUIDs)
	}()

	return &DeleteAlbumResponse{}
}

type DeleteAlbumRequest struct {
	Id uint `json:"id"`
}

type DeleteAlbumResponse struct {
}
