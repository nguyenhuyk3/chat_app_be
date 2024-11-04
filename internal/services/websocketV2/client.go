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

func NewClient(conn *websocket.Conn,
	userServices *user.UserServices,
	messageBoxId, userId, fullName string) *Client {
	return &Client{
		Conn:         conn,
		UserServices: userServices,
		Message:      make(chan *models.Message),
		MessageBoxId: messageBoxId,
		UserId:       userId,
		FullName:     fullName,
	}
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

	var incomingData struct {
		SenderId     string `json:"senderId"`
		Token        string `json:"token"`
		ReceiverId   string `json:"receiverId"`
		MessageBoxId string `json:"messageBoxId"`
		SendedId     string `json:"sendedId,omitempty"`
		Type         string `json:"type"`
		Content      string `json:"content"`
	}

	for {
		_, content, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error (ReadMessage): %v", err)
			}
			break
		}

		if err := json.Unmarshal(content, &incomingData); err != nil {
			log.Printf("error unmarshalling message: %v", err)
			continue
		}

		sendedId := incomingData.SendedId
		if incomingData.Type == "text" {
			sendedId = ""
		}

		commingMessage := &models.CommingMessage{
			MessageBoxId: incomingData.MessageBoxId,
			SenderId:     incomingData.SenderId,
			TokenDevice:  incomingData.Token,
			ReceiverId:   incomingData.ReceiverId,
			Type:         incomingData.Type,
			Content:      incomingData.Content,
			SendedId:     sendedId,
			State:        "chưa đọc",
			CreatedAt:    time.Now(),
		}
		hub.BroadcastMessage <- commingMessage
		hub.CommingMessage <- commingMessage
	}
}
