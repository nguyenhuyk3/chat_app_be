package router

import (
	"be_chat_app/api"
	"be_chat_app/internal/services/notification"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitNotificationRouter(r *gin.Engine, notificationServices *notification.NotificationServices) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	notificationApi := api.NewNotificationApi(notificationServices)

	r.GET("/notifications/get_token_by_user_id", notificationApi.GetTokenByUserId)
	r.POST("/notifications/save_token", notificationApi.SaveToken)
	r.POST("/notifications/delete_token", notificationApi.DeleteToken)
}
