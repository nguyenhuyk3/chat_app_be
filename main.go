package main

import (
	"be_chat_app/cmd"
	"be_chat_app/internal/services/notification"
	"be_chat_app/internal/services/user"
	websocketv2 "be_chat_app/internal/services/websocketV2"
	appRouter "be_chat_app/router"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	_, fireStoreClient, messagingClient, err := cmd.InitFirebase()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v\n", err)
	}
	defer cmd.CloseFirestoreClient()

	router := gin.Default()

	userServices := user.NewUserServices(fireStoreClient)
	notificationServices := notification.NewNotificatioinServices(fireStoreClient, messagingClient)
	hub := websocketv2.NewHub(messagingClient, userServices, notificationServices)
	webSocketServices := websocketv2.NewWebsocketService(hub, fireStoreClient)

	go hub.Run()
	go webSocketServices.ProcessCommingMessages()
	go webSocketServices.FetchAllMessageBoxes()

	appRouter.InitWebsocketV2Router(router, webSocketServices, userServices)
	appRouter.InitUserRouter(router, userServices)
	appRouter.InitNotificationRouter(router, notificationServices)
	appRouter.InitFileRouter(router)

	router.Static("/assets", "./assets")

	router.Run(":8080")
}
