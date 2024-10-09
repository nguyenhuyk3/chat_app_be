package websocketv2

import (
	"be_chat_app/internal/services/user"
	"be_chat_app/models"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn         *websocket.Conn
	UserServices *user.UserServices
	Message      chan *models.Message
	MessageBoxId string `json:"messageBoxId"`
	UserId       string `json:"userId"`
}

type MessageBox struct {
	MessageBoxId string `json:"messageBoxId"`
	Clients      map[string]*Client
}

func (c *Client) WriteMessage() {
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

func (c *Client) ReadMessage(hub *Hub) {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, content, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var incomingData struct {
			SenderId     string `json:"senderId"`
			MessageBoxId string `json:"messageBoxId"`
			Content      string `json:"content"`
		}

		if err := json.Unmarshal(content, &incomingData); err != nil {
			log.Printf("error: %v", err)
			continue
		}

		// commingMessage is sent to another goroutine to save message
		commingMessage := &models.CommingMessage{
			MessageBoxId: incomingData.MessageBoxId,
			SenderId:     incomingData.SenderId,
			Content:      incomingData.Content,
			State:        "Chưa đọc",
			CreatedAt:    time.Now(),
		}

		// message will be show for client
		// message := &models.Message{
		// 	SenderId:  commingMessage.SenderId,
		// 	Content:   commingMessage.Content,
		// 	State:     "chưa đọc",
		// 	CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		// }

		// hub.Broadcast <- message
		hub.Broadcast <- commingMessage
		hub.CommingMessage <- commingMessage
	}
}
