package websocket

import (
	"be_chat_app/models"
)

type RoomInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Room struct {
	RoomInfo
	Clients map[string]*Client `json:"clients"`
}

type MasterRoom struct {
	Client *Client `json:"client"`
}

func NewMasterRoom(client *Client) *MasterRoom {
	return &MasterRoom{
		Client: client,
	}
}

type Hub struct {
	MasterRooms           []*MasterRoom
	ClientOnMasterRoom    chan *Client
	Rooms                 map[string]*Room
	RoomInfo              chan *RoomInfo
	Register              chan *Client
	Unregister            chan *Client
	Broadcast             chan *Message
	BroadcastOnMasterRoom chan *Message
	MakingFriend          chan *models.FriendRequest
}

func NewHub() *Hub {
	return &Hub{
		MasterRooms:           make([]*MasterRoom, 0),
		ClientOnMasterRoom:    make(chan *Client),
		Rooms:                 make(map[string]*Room),
		RoomInfo:              make(chan *RoomInfo),
		Register:              make(chan *Client),
		Unregister:            make(chan *Client),
		Broadcast:             make(chan *Message, 5),
		BroadcastOnMasterRoom: make(chan *Message, 5),
		MakingFriend:          make(chan *models.FriendRequest, 5),
	}
}

func (h *Hub) Run() {
	for {
		select {
		// * When having the client register into the room
		case client := <-h.Register:
			// * Check if room have this client in the room
			if _, ok := h.Rooms[client.RoomID]; ok {
				// * Get a room when client register into
				r := h.Rooms[client.RoomID]
				// * Check if the new registered client is in the room
				// * If not then add this client to the room
				if _, ok := r.Clients[client.ID]; !ok {
					r.Clients[client.ID] = client
				}

			}
		// * When having the client unregister the room
		case client := <-h.Unregister:
			// * Check if the room have this client in the room
			if _, ok := h.Rooms[client.RoomID]; ok {
				// * Get the client in the room
				if _, ok := h.Rooms[client.RoomID].Clients[client.ID]; ok {
					// * If there are still members in the room
					// * then sending this message to the rest clients
					if len(h.Rooms[client.RoomID].Clients) != 0 {
						h.Broadcast <- &Message{
							Content:  "User left the room",
							RoomID:   client.RoomID,
							Username: client.Username,
						}
					}

					delete(h.Rooms[client.RoomID].Clients, client.ID)
					close(client.Message)
				}
			}
		// * When having a message is sent into the room then this case will be enabled
		case message := <-h.Broadcast:
			if _, ok := h.Rooms[message.RoomID]; ok {
				for _, cl := range h.Rooms[message.RoomID].Clients {
					cl.Message <- message
				}
			}
			// * Client go to the home page then this case will be enabled
		case clientOnMasterRoom := <-h.ClientOnMasterRoom:
			client := NewMasterRoom(clientOnMasterRoom)
			h.MasterRooms = append(h.MasterRooms, client)
			// * When the room is created then server will broast message for all clients in the master rooms
		case roomInfo := <-h.RoomInfo:
			// if _, ok := h.Rooms[roomInfo.ID]; ok {
			// 	for key := range h.Rooms {
			// 		if key == roomInfo.ID {
			// 			continue
			// 		}

			// 		if len(h.Rooms[key].Clients) != 0 {
			// 			for _, client := range h.Rooms[key].Clients {
			// 				notification := Message{
			// 					Content:  "Server just have created new room",
			// 					RoomID:   roomInfo.ID,
			// 					Username: "Server",
			// 				}

			// 				client.Message <- &notification
			// 			}
			// 		}
			// 	}
			// }
			if len(h.MasterRooms) != 0 {

				notification := Message{
					Content:  "Server just have created new room",
					RoomID:   roomInfo.ID,
					Username: "Server",
				}

				h.BroadcastOnMasterRoom <- &notification
			}
		case messageOnMasterRoom := <-h.BroadcastOnMasterRoom:
			for key := range h.MasterRooms {
				h.MasterRooms[key].Client.Message <- messageOnMasterRoom
			}
		}
	}
}
