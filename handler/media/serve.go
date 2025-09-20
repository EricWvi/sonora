package media

import (
	"net/http"
	"path/filepath"

	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/log"
	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/service"
	"github.com/gin-gonic/gin"
)

func Serve(c *gin.Context) {
	link := c.Param("link")
	m := &model.Media{}
	err := m.Get(config.ContextDB(log.MediaCtx), gin.H{
		model.Media_Link: link,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	// Initialize storage service to get file path
	storage, err := service.InitLocalStorageService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to initialize storage service"})
		return
	}

	// Check if file exists
	if !storage.FileExists(m.Key) {
		c.JSON(http.StatusNotFound, gin.H{"message": "File not found"})
		return
	}

	// Get full file path
	fullPath := storage.GetFilePath(m.Key)
	
	// Extract filename from the key for Content-Disposition header
	filename := filepath.Base(m.Key)
	
	// Serve the file directly
	c.Header("Content-Disposition", "inline; filename=\""+filename+"\"")
	c.File(fullPath)
}
