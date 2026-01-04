package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func CreateTestMedia(t *testing.T, router *gin.Engine) uuid.UUID {
	// Create a multipart form with a test file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("photos", "test.txt")
	assert.NoError(t, err)

	_, err = part.Write([]byte("test media content"))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	// Upload the file
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Photos []string `json:"photos"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Photos, 1)

	mediaUUID, err := uuid.Parse(resp.Photos[0])
	assert.NoError(t, err)

	return mediaUUID
}
