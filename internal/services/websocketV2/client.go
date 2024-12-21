package websocketv2

import (
	"be_chat_app/internal/services/user"
	"be_chat_app/models"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Client struct {
	Conn         *websocket.Conn
	UserServices *user.UserServices
	Message      chan *models.Message
	MessageBoxId string `json:"messageBoxId"`
	UserId       string `json:"userId"`
	FullName     string `json:"fullName"`
	// StartTime    *time.Time
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
		// StartTime:    nil,
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

func (c *Client) ReadMessage(h *Hub) {
	defer func() {
		h.ClientGetOutMessageBox <- c
		c.Conn.Close()
	}()

	for {
		_, content, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error (ReadMessage): %v", err)
			}
			break
		}

		var incomingData struct {
			SenderId     string                     `json:"senderId"`
			Token        string                     `json:"token"`
			ReceiverId   string                     `json:"receiverId"`
			MessageBoxId string                     `json:"messageBoxId"`
			SendedId     string                     `json:"sendedId,omitempty"`
			Type         string                     `json:"type"`
			CallType     string                     `json:"callType,omitempty"`
			Content      string                     `json:"content"`
			Sdp          *webrtc.SessionDescription `json:"sdp,omitempty"`
			Candidate    *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
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
			CallType:     incomingData.CallType,
			Type:         incomingData.Type,
			Content:      incomingData.Content,
			Sdp:          incomingData.Sdp,
			Candidate:    incomingData.Candidate,
			SendedId:     sendedId,
			State:        "chưa đọc",
			CreatedAt:    time.Now(),
		}

		switch incomingData.Type {
		case "video", "audio", "text", "completed-media-call", "missed-media-call", "declined-media-call":
			h.BroadcastMessage <- commingMessage
			h.CommingMessage <- commingMessage
		case "offer", "answer", "ice-candidate":
			// if incomingData.Type == "answer" {
			// 	if client, ok := h.MessageBoxes[incomingData.MessageBoxId].Clients[incomingData.ReceiverId]; ok {
			// 		if client.StartTime == nil {
			// 			currentTime := time.Now()
			// 			client.StartTime = &currentTime
			// 			fmt.Println(currentTime)
			// 		}
			// 	}
			// 	if client, ok := h.MessageBoxes[incomingData.MessageBoxId].Clients[incomingData.SenderId]; ok {
			// 		if client.StartTime == nil {
			// 			currentTime := time.Now()
			// 			client.StartTime = &currentTime
			// 			fmt.Println(currentTime)
			// 		}
			// 	}
			// }
			h.BroadcastMessage <- commingMessage
		case "completed-media-call-signal", "missed-media-call-signal", "declined-media-call-signal":
			// if client, ok := h.MessageBoxes[incomingData.MessageBoxId].Clients[incomingData.ReceiverId]; ok {
			// 	if client.StartTime != nil {
			// 		duration := time.Since(*client.StartTime)
			// 		fmt.Println(*client.StartTime)
			// 		fmt.Println(duration)
			// 		fmt.Println(incomingData.ReceiverId)
			// 		fmt.Println("=====================")
			// 		client.StartTime = nil
			// 	}
			// }
			// if client, ok := h.MessageBoxes[incomingData.MessageBoxId].Clients[incomingData.SenderId]; ok {
			// 	if client.StartTime != nil {
			// 		duration := time.Since(*client.StartTime)
			// 		fmt.Println(*client.StartTime)
			// 		fmt.Println(duration)
			// 		fmt.Println(incomingData.SenderId)
			// 		fmt.Println("=====================")
			// 		client.StartTime = nil
			// 	}
			// }
			h.BroadcastMessage <- commingMessage
		case "toggle-action":
			h.BroadcastMessage <- commingMessage
		}
	}
}
