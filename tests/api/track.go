package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler/track"
	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/tests"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func CreateTestTrack(t *testing.T, router *gin.Engine, trackField model.TrackField) uint {
	createReq := track.CreateTrackRequest{
		TrackField: trackField,
	}
	reqBody, _ := json.Marshal(createReq)
	code, respBody := tests.ServeJSON(router, http.MethodPost, "/track?Action=CreateTrack", reqBody)
	assert.Equal(t, http.StatusOK, code)

	var resp struct {
		Message track.CreateTrackResponse `json:"message"`
	}
	err := json.Unmarshal(respBody, &resp)
	assert.NoError(t, err)
	assert.NotZero(t, resp.Message.Id)

	return resp.Message.Id
}

func DeleteTestTrack(t *testing.T, router *gin.Engine, id uint) {
	// Check if track still exists first
	db := config.DB()
	var count int64
	db.Model(&model.Track{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count)
	if count == 0 {
		// Track already deleted, skip
		return
	}

	deleteReq := track.DeleteTrackRequest{
		Id: id,
	}
	reqBody, _ := json.Marshal(deleteReq)
	code, respBody := tests.ServeJSON(router, http.MethodPost, "/track?Action=DeleteTrack", reqBody)
	assert.Equal(t, http.StatusOK, code)

	var resp struct {
		Message track.DeleteTrackResponse `json:"message"`
	}
	err := json.Unmarshal(respBody, &resp)
	assert.NoError(t, err)
}
