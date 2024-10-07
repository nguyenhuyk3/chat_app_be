package websocketv2

import (
	"be_chat_app/internal/services/user"
	"be_chat_app/models"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn         *websocket.Conn
	UserServices *user.UserServices
	Message      chan *models.Message
	UserId       string `json:"userId"`
}

func (c *Client) writeMessage() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Message
		if !ok {
			return
		}

		c.Conn.WriteJSON(message)
	}
}

func (c *Client) readMessage(hub *Hub) {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

	}
}
