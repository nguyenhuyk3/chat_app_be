package router

import (
	"be_chat_app/services/websocket"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func InitRouter(wsHandler *websocket.Handler) {
	r = gin.Default()

	r.POST("/ws/createRoom", wsHandler.CreateRoom)
	r.GET("/ws/joinRoom/:roomId", wsHandler.JoinRoom)
	r.GET("/ws/getRooms", wsHandler.GetRooms)
	r.GET("/ws/getClients/:roomId", wsHandler.GetClients)
}

func Start(addr string) error {
	return r.Run(addr)
}
