package router

import (
	"be_chat_app/internal/services/websocket"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitWebsocketRouter(r *gin.Engine, ws *websocket.WebsocketService) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/ws/createRoom", ws.CreateRoom)
	r.GET("/ws/joinMasterRoom", ws.JoinMasterRoom)
	r.GET("/ws/joinRoom/:roomId", ws.JoinRoom)
	r.GET("/ws/getRooms", ws.GetRooms)
	r.GET("/ws/getClients/:roomId", ws.GetClients)
}
