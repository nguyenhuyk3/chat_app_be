package websocket

type Room struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Clients map[string]*Client `json:"clients"`
}

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan *Message, 5),
	}
}

func (h *Hub) Run() {
	for {
		select {
		// * When having the client register into the room
		case cl := <-h.Register:
			// * Check if room have this client in the room
			if _, ok := h.Rooms[cl.RoomID]; ok {
				// * Get a room when client register into
				r := h.Rooms[cl.RoomID]
				// * Check if the new registered client is in the room
				// * If not then add this client to the room
				if _, ok := r.Clients[cl.ID]; !ok {
					r.Clients[cl.ID] = cl
				}

			}
		// * When having the client unregister the room
		case cl := <-h.Unregister:
			// * Check if the room have this client in the room
			if _, ok := h.Rooms[cl.RoomID]; ok {
				// * Get the client in the room
				if _, ok := h.Rooms[cl.RoomID].Clients[cl.ID]; ok {
					// * If there are still members in the room
					// * then sending this message to the rest clients
					if len(h.Rooms[cl.RoomID].Clients) != 0 {
						h.Broadcast <- &Message{
							Content:  "User left the room",
							RoomID:   cl.RoomID,
							Username: cl.Username,
						}
					}

					delete(h.Rooms[cl.RoomID].Clients, cl.ID)
					close(cl.Message)
				}
			}
		case m := <-h.Broadcast:
			if _, ok := h.Rooms[m.RoomID]; ok {
				for _, cl := range h.Rooms[m.RoomID].Clients {
					cl.Message <- m
				}
			}
		}
	}
}
