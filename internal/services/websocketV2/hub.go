package websocketv2

import (
	"be_chat_app/models"
)

type Hub struct {
	MasterRooms              map[string]*MasterRoom
	ClientGetInToMasterRoom  chan *ClientOnMasterRoom
	ClientGetOutMasterRoom   chan *ClientOnMasterRoom
	AcceptFriendNotification chan *models.Notification
	CommingMessage           chan *models.CommingMessage
	Broadcast                chan *models.CommingMessage
	ClientGetInMessageBox    chan *Client
	ClientGetOutMessageBox   chan *Client
	MessageBoxes             map[string]*MessageBox
}

func NewHub() *Hub {
	return &Hub{
		MasterRooms:              make(map[string]*MasterRoom),
		ClientGetInToMasterRoom:  make(chan *ClientOnMasterRoom),
		ClientGetOutMasterRoom:   make(chan *ClientOnMasterRoom),
		AcceptFriendNotification: make(chan *models.Notification, 2),
		CommingMessage:           make(chan *models.CommingMessage),
		Broadcast:                make(chan *models.CommingMessage),
		ClientGetInMessageBox:    make(chan *Client),
		ClientGetOutMessageBox:   make(chan *Client),
		MessageBoxes:             make(map[string]*MessageBox),
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
			}
		case message := <-h.Broadcast:
			messageBoxId := message.MessageBoxId

			if _, ok := h.MessageBoxes[messageBoxId]; ok {
				newMessage := &models.Message{
					SenderId:  message.SenderId,
					Content:   message.Content,
					State:     message.State,
					CreatedAt: message.CreatedAt.String(),
				}
				for _, client := range h.MessageBoxes[messageBoxId].Clients {
					client.Message <- newMessage
				}
			}
		}
	}
}
