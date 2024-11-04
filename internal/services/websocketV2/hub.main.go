package websocketv2

import (
	"be_chat_app/internal/services/notification"
	"be_chat_app/internal/services/user"
	"be_chat_app/models"

	"firebase.google.com/go/messaging"
)

type Hub struct {
	MessagingClient      *messaging.Client
	UserServices         *user.UserServices
	NotificationServices *notification.NotificationServices
	// When client join home then this place they will join first
	MasterRooms             map[string]*MasterRoom
	ClientGetInToMasterRoom chan *ClientOnMasterRoom
	ClientGetOutMasterRoom  chan *ClientOnMasterRoom

	ReadedMessageNotification chan *models.CommingMessage
	AcceptFriendNotification  chan *models.Notification

	CommingMessage   chan *models.CommingMessage
	BroadcastMessage chan *models.CommingMessage

	ClientGetInMessageBox  chan *Client
	ClientGetOutMessageBox chan *Client
	// This will manage all message box between two clients
	MessageBoxes map[string]*MessageBox
}

func NewHub(messagingClient *messaging.Client, userServices *user.UserServices, notificationServices *notification.NotificationServices) *Hub {
	return &Hub{
		MessagingClient:           messagingClient,
		UserServices:              userServices,
		NotificationServices:      notificationServices,
		MasterRooms:               make(map[string]*MasterRoom),
		ClientGetInToMasterRoom:   make(chan *ClientOnMasterRoom),
		ClientGetOutMasterRoom:    make(chan *ClientOnMasterRoom, 10),
		ReadedMessageNotification: make(chan *models.CommingMessage, 10),
		AcceptFriendNotification:  make(chan *models.Notification, 2),
		CommingMessage:            make(chan *models.CommingMessage),
		BroadcastMessage:          make(chan *models.CommingMessage),
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

		case message := <-h.BroadcastMessage:
			h.handleBroadcastMessage(message)

		case readedMessage := <-h.ReadedMessageNotification:
			h.handleReadedMessageNotification(readedMessage)
		}
	}
}
