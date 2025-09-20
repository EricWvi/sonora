package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LocalStorageService handles local file storage operations
type LocalStorageService struct {
	BasePath string
}

// InitLocalStorageService initializes a new local storage service
func InitLocalStorageService() (*LocalStorageService, error) {
	basePath := "objects"

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorageService{
		BasePath: basePath,
	}, nil
}

// UploadMultipartFile saves a multipart file to local storage using provided UUID
// Returns the relative path as the key
func (s *LocalStorageService) UploadMultipartFile(c *gin.Context, file *multipart.FileHeader, fileUUID uuid.UUID) (string, error) {
	// Use provided UUID for directory structure
	fileID := fileUUID.String()

	// Get first 4 characters for directory structure
	if len(fileID) < 4 {
		return "", fmt.Errorf("provided UUID is too short")
	}

	firstTwo := fileID[:2]
	secondTwo := fileID[2:4]

	// Create directory structure: objects/<ab>/<ab>/
	dirPath := filepath.Join(s.BasePath, firstTwo, secondTwo)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Use original filename
	filename := file.Filename

	// Full path for saving
	fullPath := filepath.Join(dirPath, filename)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Return relative path as key
	relativePath := filepath.Join(firstTwo, secondTwo, filename)
	return relativePath, nil
}

// DeleteFile removes a file from local storage
func (s *LocalStorageService) DeleteFile(key string) error {
	fullPath := filepath.Join(s.BasePath, key)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", key)
	}

	// Remove the file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFilePath returns the full path to a file given its key
func (s *LocalStorageService) GetFilePath(key string) string {
	return filepath.Join(s.BasePath, key)
}

// FileExists checks if a file exists in storage
func (s *LocalStorageService) FileExists(key string) bool {
	fullPath := filepath.Join(s.BasePath, key)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}

// DeleteObject removes a file from storage (alias for DeleteFile)
func (s *LocalStorageService) DeleteObject(c *gin.Context, key string) error {
	return s.DeleteFile(key)
}
