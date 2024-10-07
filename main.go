package main

import (
	"be_chat_app/cmd"
	"be_chat_app/internal/services/user"
	websocketv2 "be_chat_app/internal/services/websocketV2"
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

	hub := websocketv2.NewHub()
	userServices := user.NewUserServices(client)
	webSocketServices := websocketv2.NewWebsocketService(hub)

	go hub.Run()

	appRouter.InitWebsocketV2Router(router, webSocketServices, userServices)
	appRouter.InitUserRouter(router, userServices)

	router.Run(":8080")
}
