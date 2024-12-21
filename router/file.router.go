package router

import (
	"be_chat_app/api"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitFileRouter(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	fileApi := api.NewFileApi()

	r.POST("/file/upload_video_file", fileApi.UploadVideoFile)
	r.POST("/file/upload_audio_file", fileApi.UploadAudioFile)

}
