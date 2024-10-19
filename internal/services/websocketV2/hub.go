package websocketv2

import (
	"be_chat_app/internal/services/notification"
	"be_chat_app/models"
	"time"

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
		ClientGetInMessageBox:     make(chan *Client),
		ClientGetOutMessageBox:    make(chan *Client),
		MessageBoxes:              make(map[string]*MessageBox),
	}
}

func (h *Hub) Run() {
	for {
		select {
		// When client get into master room
		case clientGetIntoMasterRoom := <-h.ClientGetInToMasterRoom:
			userId := clientGetIntoMasterRoom.UserId

			if _, ok := h.MasterRooms[userId]; !ok {
				newMasterRoom := NewMasterRoom(clientGetIntoMasterRoom)

				h.MasterRooms[userId] = newMasterRoom
			}
		// When client get out master room
		case clientGetOutMasterRoom := <-h.ClientGetOutMasterRoom:
			userId := clientGetOutMasterRoom.UserId

			if _, ok := h.MasterRooms[userId]; ok {
				delete(h.MasterRooms, userId)
				close(clientGetOutMasterRoom.AcceptFriendNotification)
			}
		case acceptFriendNotification := <-h.AcceptFriendNotification:
			userId := acceptFriendNotification.ToUserId

			if _, ok := h.MasterRooms[userId]; ok {
				h.MasterRooms[userId].ClientOnMasterRoom.AcceptFriendNotification <- acceptFriendNotification
			}
		case clientGetInMessageBox := <-h.ClientGetInMessageBox:
			messageBoxId := clientGetInMessageBox.MessageBoxId
			userId := clientGetInMessageBox.UserId

			if _, ok := h.MessageBoxes[messageBoxId].Clients[userId]; !ok {
				h.MessageBoxes[messageBoxId].Clients[userId] = clientGetInMessageBox

				if len(h.MessageBoxes[messageBoxId].Clients) == 2 {
					for key := range h.MessageBoxes[messageBoxId].Clients {
						if key != userId {
							newMessage := &models.CommingMessage{
								SenderId:     userId,
								MessageBoxId: messageBoxId,
								Content:      "anhiuemlove33333!@#@#@!!!****&(*&@(^&*()concak",
								State:        "",
							}

							h.ReadedMessageNotification <- newMessage
							break
						}
					}
				}
			}
		case clienGetOutMessageBox := <-h.ClientGetOutMessageBox:
			messageBoxId := clienGetOutMessageBox.MessageBoxId
			userId := clienGetOutMessageBox.UserId

			if _, ok := h.MessageBoxes[messageBoxId].Clients[userId]; ok {
				delete(h.MessageBoxes[messageBoxId].Clients, userId)
				close(clienGetOutMessageBox.Message)
			}
		case message := <-h.Broadcast:
			messageBoxId := message.MessageBoxId

			if _, ok := h.MessageBoxes[messageBoxId]; ok {
				newState := "chưa đọc"
				if len(h.MessageBoxes[messageBoxId].Clients) == 2 {
					newState = "đã đọc"
				}

				newMessage := &models.Message{
					SenderId:  message.SenderId,
					Content:   message.Content,
					State:     newState,
					CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
				}
				for _, client := range h.MessageBoxes[messageBoxId].Clients {
					if len(h.MessageBoxes[messageBoxId].Clients) == 1 {
						newNotification := notification.MessageNotification{
							Token:  message.TokenDevice,
							Avatar: "",
							Title:  h.MessageBoxes[messageBoxId].Clients[message.SenderId].FullName,
							Body:   "1 tin nhắn mới",
						}

						go func() {
							notification.SendNotificationForCommingMessage(h.MessagingClient, newNotification)
						}()
					}
					client.Message <- newMessage
				}
			}
		case readedMessage := <-h.ReadedMessageNotification:
			userId := readedMessage.SenderId
			messageBoxId := readedMessage.MessageBoxId
			content := readedMessage.Content

			if _, ok := h.MessageBoxes[messageBoxId]; ok {
				newMessage := &models.Message{
					SenderId:  userId,
					Content:   content,
					State:     "",
					CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
				}

				for _, client := range h.MessageBoxes[messageBoxId].Clients {
					client.Message <- newMessage
				}
			}
		}
	}
}
