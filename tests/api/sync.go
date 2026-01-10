package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/EricWvi/sonora/handler/sync"
	"github.com/EricWvi/sonora/tests"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func GetFullSync(t *testing.T, router *gin.Engine) *sync.GetFullSyncResponse {
	code, respBody := tests.ServeJSON(router, http.MethodGet, "/sync?Action=GetFullSync", nil)
	assert.Equal(t, http.StatusOK, code)

	var resp struct {
		Message sync.GetFullSyncResponse `json:"message"`
	}
	err := json.Unmarshal(respBody, &resp)
	assert.NoError(t, err)

	return &resp.Message
}

func GetUpdates(t *testing.T, router *gin.Engine, since int64) *sync.GetUpdatesResponse {
	url := fmt.Sprintf("/sync?Action=GetUpdates&since=%d", since)
	code, respBody := tests.ServeJSON(router, http.MethodGet, url, nil)
	assert.Equal(t, http.StatusOK, code)

	var resp struct {
		Message sync.GetUpdatesResponse `json:"message"`
	}
	err := json.Unmarshal(respBody, &resp)
	assert.NoError(t, err)

	return &resp.Message
}
