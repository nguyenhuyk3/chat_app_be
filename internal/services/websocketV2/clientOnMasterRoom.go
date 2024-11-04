package websocketv2

import (
	"be_chat_app/models"
	"encoding/json"
	"log"

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

type ClientOnMasterRoom struct {
	Conn                               *websocket.Conn
	PeerConnection                     *webrtc.PeerConnection
	AcceptFriendNotification           chan *models.Notification
	UserStatus                         chan *UserStatus
	LastStateForMessageBoxOnMasterRoom chan *LastStateForMessageBoxOnMasterRoom
	// DeviceToken              string `json:"deviceToken"`
	UserId string `json:"userId"`
	// UserName string `json:"userName"`
}

func NewClientOnMasterRoom(
	conn *websocket.Conn,
	userId string) *ClientOnMasterRoom {
	return &ClientOnMasterRoom{
		Conn:                               conn,
		AcceptFriendNotification:           make(chan *models.Notification),
		UserStatus:                         make(chan *UserStatus, 2),
		LastStateForMessageBoxOnMasterRoom: make(chan *LastStateForMessageBoxOnMasterRoom, 2),
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
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error (ReadMessages [ClientOnMasterRoom]): %v", err)
			}
			break
		}
	}
}

type Message interface{}

func (c *ClientOnMasterRoom) sendMessage(message Message) {
	// Encode the message as JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}
	// Send message via WebSocket
	err = c.Conn.WriteMessage(websocket.TextMessage, messageJSON)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}

type VideoMessage struct {
	Type         string                     `json:"type"`
	ICECandidate *webrtc.ICECandidateInit   `json:"ice_candidate,omitempty"`
	SessionDesc  *webrtc.SessionDescription `json:"session_description,omitempty"`
}

// Gửi ICE Candidate
func (c *ClientOnMasterRoom) SendICECandidate(candidate webrtc.ICECandidateInit) {
	// Send candidate via WebSocket
	message := VideoMessage{
		Type:         "ice_candidate",
		ICECandidate: &candidate,
	}
	c.sendMessage(message)
}

// Send SDP Offer
func (c *ClientOnMasterRoom) SendCallOffer(offer webrtc.SessionDescription) {
	message := VideoMessage{
		Type:        "call_offer",
		SessionDesc: &offer,
	}
	c.sendMessage(message)
}

func (c *ClientOnMasterRoom) SendSDP(sdp string) {
	// Create message to send SDP
	message := VideoMessage{
		Type: "sdp_answer",
		SessionDesc: &webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  sdp,
		},
	}
	c.sendMessage(message)
}

func handleICECandidate(candidate webrtc.ICECandidateInit, pc *webrtc.PeerConnection) {
	// Add ICE candidate to Peer Connection
	err := pc.AddICECandidate(candidate)
	if err != nil {
		log.Println("Error adding ICE candidate:", err)
	}
}
