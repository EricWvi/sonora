package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) SearchTrack(c *gin.Context, req *SearchTrackRequest) *SearchTrackResponse {
	if req.Query == "" {
		handler.Errorf(c, "query parameter is required")
		return nil
	}

	var where model.WhereExpr
	where.ILIKE(model.Track_Name, "%"+req.Query+"%")

	tracks, err := model.ListAllTracks(config.ContextDB(c), where)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &SearchTrackResponse{
		Tracks: tracks,
	}
}

type SearchTrackRequest struct {
	Query string `form:"query"`
}

type SearchTrackResponse struct {
	Tracks []model.TrackView `json:"tracks"`
}