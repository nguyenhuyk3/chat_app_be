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
	FullName     string `json:"fullName"`
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
		hub.ClientGetOutMessageBox <- c
		c.Conn.Close()
	}()

	for {
		_, content, err := c.Conn.ReadMessage()
		if err != nil {
			/*
				websocket.CloseGoingAway: status code indicating that the connection is being closed for some reason
					(e.g. the client or server is leaving).
				websocket.CloseAbnormalClosure: status code indicating that the connection was closed abnormally.
			*/
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error (ReadMessage): %v", err)
			}
			break
		}

		var incomingData struct {
			SenderId     string `json:"senderId"`
			Token        string `json:"token"`
			ReceiverId   string `json:"receiverId"`
			MessageBoxId string `json:"messageBoxId"`
			Content      string `json:"content"`
		}

		if err := json.Unmarshal(content, &incomingData); err != nil {
			log.Printf("error when reading message: %v", err)
			continue
		}

		// commingMessage is sent to another goroutine to save message
		commingMessage := &models.CommingMessage{
			MessageBoxId: incomingData.MessageBoxId,
			SenderId:     incomingData.SenderId,
			TokenDevice:  incomingData.Token,
			ReceiverId:   incomingData.ReceiverId,
			Content:      incomingData.Content,
			State:        "chưa đọc",
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
