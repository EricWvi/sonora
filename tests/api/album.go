package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/EricWvi/sonora/handler/album"
	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/tests"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func CreateTestAlbum(t *testing.T, router *gin.Engine, albumField model.AlbumField) uint {
	createReq := album.CreateAlbumRequest{
		AlbumField: albumField,
	}
	reqBody, _ := json.Marshal(createReq)
	code, respBody := tests.ServeJSON(router, http.MethodPost, "/album?Action=CreateAlbum", reqBody)
	assert.Equal(t, http.StatusOK, code)

	var resp struct {
		Message album.CreateAlbumResponse `json:"message"`
	}
	err := json.Unmarshal(respBody, &resp)
	assert.NoError(t, err)
	assert.NotZero(t, resp.Message.Id)

	return resp.Message.Id
}

func DeleteTestAlbum(t *testing.T, router *gin.Engine, id uint) {
	createReq := album.DeleteAlbumRequest{
		Id: id,
	}
	reqBody, _ := json.Marshal(createReq)
	code, respBody := tests.ServeJSON(router, http.MethodPost, "/album?Action=DeleteAlbum", reqBody)
	assert.Equal(t, http.StatusOK, code)

	var resp struct {
		Message album.DeleteAlbumResponse `json:"message"`
	}
	err := json.Unmarshal(respBody, &resp)
	assert.NoError(t, err)
}
