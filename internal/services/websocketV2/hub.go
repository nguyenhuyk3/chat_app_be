package websocketv2

import (
	"be_chat_app/models"
)

type MasterRoom struct {
	ClientOnMasterRoom *ClientOnMasterRoom `json:"clientOnMasterRoom"`
}

func NewMasterRoom(client *ClientOnMasterRoom) *MasterRoom {
	return &MasterRoom{
		ClientOnMasterRoom: client,
	}
}

type Hub struct {
	MasterRooms              map[string]*MasterRoom
	ClientGetInToMasterRoom  chan *ClientOnMasterRoom
	ClientGetOutMasterRoom   chan *ClientOnMasterRoom
	AcceptFriendNotification chan *models.Notification
}

func NewHub() *Hub {
	return &Hub{
		MasterRooms:              make(map[string]*MasterRoom),
		ClientGetInToMasterRoom:  make(chan *ClientOnMasterRoom),
		ClientGetOutMasterRoom:   make(chan *ClientOnMasterRoom),
		AcceptFriendNotification: make(chan *models.Notification, 2),
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
		}
	}
}
