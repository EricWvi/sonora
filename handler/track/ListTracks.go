package track

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) ListTracks(c *gin.Context, req *ListTracksRequest) *ListTracksResponse {
	m := model.WhereExpr{}

	tracks, hasMore, err := model.ListTracks(config.ContextDB(c), m, req.Page)
	if err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &ListTracksResponse{
		Tracks:  tracks,
		HasMore: hasMore,
	}
}

type ListTracksRequest struct {
	Page      uint   `form:"page"`
	Condition string `form:"condition"`
}

type QueryCondition struct {
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type ListTracksResponse struct {
	Tracks  []model.TrackView `json:"tracks"`
	HasMore bool              `json:"hasMore"`
}