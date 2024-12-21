package websocketv2

import (
	"be_chat_app/models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type UserStatus struct {
	UserId      string `json:"userId"`
	IsOnline    bool   `json:"isOnline"`
	TokenDevice string `json:"tokenDevice"`
}

type LastStateForMessageBoxOnMasterRoom struct {
	IsLastStateStruct bool   `json:"isLastStateStruct"`
	MessageBoxId      string `json:"messageBoxId"`
	SenderId          string `json:"senderId"`
	LastMessage       string `json:"lastMessage"`
	LastTime          string `json:"lastTime"`
	LastStatus        string `json:"lastStatus"`
}

type OfferNotification struct {
	MessageBoxId    string                     `json:"messageBoxId"`
	SenderId        string                     `json:"senderId"`
	SenderName      string                     `json:"senderName"`
	SenderAvatarUrl string                     `json:"senderAvatarUrl,omitempty"`
	CallType        string                     `json:"callType"`
	Type            string                     `json:"type"`
	Token           string                     `json:"token"`
	Sdp             *webrtc.SessionDescription `json:"sdp,omitempty"`
	Candidate       *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
}

type ClientOnMasterRoom struct {
	Conn                               *websocket.Conn
	PeerConnection                     *webrtc.PeerConnection
	AcceptFriendNotification           chan *models.Notification
	UserStatus                         chan *UserStatus
	LastStateForMessageBoxOnMasterRoom chan *LastStateForMessageBoxOnMasterRoom
	OfferNotification                  chan *OfferNotification
	// DeviceToken              string `json:"deviceToken"`
	UserId string `json:"userId"`
	// UserName string `json:"userName"`
}

func NewClientOnMasterRoom(
	conn *websocket.Conn,
	userId string) *ClientOnMasterRoom {
	return &ClientOnMasterRoom{
		Conn:                               conn,
		PeerConnection:                     nil,
		AcceptFriendNotification:           make(chan *models.Notification),
		UserStatus:                         make(chan *UserStatus, 2),
		LastStateForMessageBoxOnMasterRoom: make(chan *LastStateForMessageBoxOnMasterRoom, 2),
		OfferNotification:                  make(chan *OfferNotification),
		UserId:                             userId,
	}
}

type MasterRoom struct {
	ClientOnMasterRoom *ClientOnMasterRoom `json:"clientOnMasterRoom"`
}

func NewMasterRoom(client *ClientOnMasterRoom) *MasterRoom {
	return &MasterRoom{
		ClientOnMasterRoom: client,
	}
}

func (c *ClientOnMasterRoom) WriteAcceptNotification() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.AcceptFriendNotification
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}
}

func (c *ClientOnMasterRoom) WriteUserStatus() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.UserStatus
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}
}

func (c *ClientOnMasterRoom) WriteLastStateForMessageBoxOnMasterRoom() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.LastStateForMessageBoxOnMasterRoom
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}
}

func (c *ClientOnMasterRoom) WriteOffer() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.OfferNotification
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}
}

func (c *ClientOnMasterRoom) SendOnlineFriends(onlineFriends []string) {
	message := map[string]interface{}{
		"type":    "onlineFriends",
		"friends": onlineFriends,
	}
	c.Conn.WriteJSON(message)
}

func (c *ClientOnMasterRoom) ReadMessages(h *Hub) {
	defer func() {
		c.Conn.Close()
		h.ClientGetOutMasterRoom <- c
	}()
	for {
		_, content, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error (ReadMessages [ClientOnMasterRoom]): %v", err)
			}
			c.Conn.Close()
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

		fmt.Println("lksjflks" + " " + incomingData.Type)

		if incomingData.Type == "answer" ||
			incomingData.Type == "ice-candidate" ||
			incomingData.Type == "declined-media-call-signal" ||
			incomingData.Type == "declined-media-call-at-foreground" {
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
				SendedId:     "",
				State:        "",
				CreatedAt:    time.Now(),
			}

			h.BroadcastMessage <- commingMessage
		}
	}
}
