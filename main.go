package main

import (
	"be_chat_app/router"
	"be_chat_app/services/websocket"
)

func main() {
	hub := websocket.NewHub()
	wsHandler := websocket.NewHandler(hub)

	go hub.Run()

	router.InitRouter(wsHandler)

	router.Start("0.0.0.0:8080")
}
