package main

import (
	"be_chat_app/cmd"
	"be_chat_app/internal/services/websocket"
	appRouter "be_chat_app/router"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	_, client, err := cmd.InitFirebase()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v\n", err)
	}
	defer cmd.CloseFirestoreClient()

	router := gin.Default()

	hub := websocket.NewHub()
	websocketService := websocket.NewWebsocketService(hub)

	go hub.Run()

	appRouter.InitWebsocketRouter(router, websocketService)
	appRouter.InitUserRouter(router, client)

	router.Run(":8080")
}
