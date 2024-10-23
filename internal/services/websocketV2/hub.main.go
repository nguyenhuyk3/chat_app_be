package websocketv2

import (
	"be_chat_app/models"

	"firebase.google.com/go/messaging"
)

type Hub struct {
	MessagingClient *messaging.Client
	// When client join home then this place they will join first
	MasterRooms               map[string]*MasterRoom
	ClientGetInToMasterRoom   chan *ClientOnMasterRoom
	ClientGetOutMasterRoom    chan *ClientOnMasterRoom
	ReadedMessageNotification chan *models.CommingMessage
	AcceptFriendNotification  chan *models.Notification
	CommingMessage            chan *models.CommingMessage
	Broadcast                 chan *models.CommingMessage
	FileBroadCast             chan *models.FileMessage
	ClientGetInMessageBox     chan *Client
	ClientGetOutMessageBox    chan *Client
	// This will manage all message box between two clients
	MessageBoxes map[string]*MessageBox
}

func NewHub(messagingClient *messaging.Client) *Hub {
	return &Hub{
		MessagingClient:           messagingClient,
		MasterRooms:               make(map[string]*MasterRoom),
		ClientGetInToMasterRoom:   make(chan *ClientOnMasterRoom),
		ClientGetOutMasterRoom:    make(chan *ClientOnMasterRoom),
		ReadedMessageNotification: make(chan *models.CommingMessage, 10),
		AcceptFriendNotification:  make(chan *models.Notification, 2),
		CommingMessage:            make(chan *models.CommingMessage),
		Broadcast:                 make(chan *models.CommingMessage),
		FileBroadCast:             make(chan *models.FileMessage, 10),
		ClientGetInMessageBox:     make(chan *Client),
		ClientGetOutMessageBox:    make(chan *Client),
		MessageBoxes:              make(map[string]*MessageBox),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case clientGetIntoMasterRoom := <-h.ClientGetInToMasterRoom:
			h.handleClientGetIntoMasterRoom(clientGetIntoMasterRoom)

		case clientGetOutMasterRoom := <-h.ClientGetOutMasterRoom:
			h.handleClientGetOutMasterRoom(clientGetOutMasterRoom)

		case acceptFriendNotification := <-h.AcceptFriendNotification:
			h.handleAcceptFriendNotification(acceptFriendNotification)

		case clientGetInMessageBox := <-h.ClientGetInMessageBox:
			h.handleClientGetInMessageBox(clientGetInMessageBox)

		case clientGetOutMessageBox := <-h.ClientGetOutMessageBox:
			h.handleClientGetOutMessageBox(clientGetOutMessageBox)

		case message := <-h.Broadcast:
			h.handleBroadcastMessage(message)

		case readedMessage := <-h.ReadedMessageNotification:
			h.handleReadedMessageNotification(readedMessage)
		}
	}
}
