package media

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/log"
	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Upload handles the media upload request from form data.
func Upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{"message": "Failed to parse multipart form: " + err.Error()})
		return
	}

	files := form.File["photos"]
	if len(files) == 0 {
		c.JSON(400, gin.H{"message": "No files found in form data"})
		return
	}

	client, err := service.InitLocalStorageService()
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	var fileIds []string
	for _, file := range files {
		// Generate UUID in Go binary first
		fileUUID := uuid.New()

		// Upload file using the generated UUID
		fileKey, err := client.UploadMultipartFile(c, file, fileUUID)
		if err != nil {
			c.JSON(500, gin.H{"message": "Failed to save " + file.Filename + ": " + err.Error()})
			return
		}

		// Create media record at the end with the UUID and file key
		m := &model.Media{
			Link: fileUUID,
			Key:  fileKey,
		}
		err = m.Create(config.ContextDB(log.MediaCtx))
		if err != nil {
			c.JSON(500, gin.H{"message": "Failed to create media record: " + err.Error()})
			return
		}

		fileIds = append(fileIds, m.Link.String())
	}

	c.JSON(200, gin.H{
		"photos": fileIds,
	})
}
