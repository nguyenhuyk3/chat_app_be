package router

import (
	"be_chat_app/api"
	"be_chat_app/internal/services/user"
	websocketv2 "be_chat_app/internal/services/websocketV2"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitWebsocketV2Router(r *gin.Engine,
	webSocketServices *websocketv2.WebsocketServices,
	userServices *user.UserServices) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	wsa := api.NewWebsocketApi(webSocketServices, userServices)

	r.GET("/ws/join_master_room", wsa.JoinMasterRoom)
	r.GET("/ws/join_message_box/:message_box_id", wsa.JoinMessageBox)
	r.POST("/ws/accept_friend", wsa.AcceptFriend)
	r.POST("/ws/logout", wsa.Logout)
}
