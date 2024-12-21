package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type FileApi struct{}

func NewFileApi() *FileApi {
	return &FileApi{}
}

func (f *FileApi) UploadVideoFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not found"})
		return
	}

	dst := filepath.Join("./assets/videos", file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"fileUrl": fmt.Sprintf("http://192.168.1.7:8080/assets/videos/%s", file.Filename)})
}

func (f *FileApi) UploadAudioFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file not found"})
		return
	}

	dst := filepath.Join("./assets/audios", file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"fileUrl": fmt.Sprintf("http://192.168.1.7:8080/assets/audios/%s", file.Filename)})
}
