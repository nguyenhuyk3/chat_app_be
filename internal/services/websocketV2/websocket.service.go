package websocketv2

type WebsocketServices struct {
	Hub *Hub
}

func NewWebsocketService(h *Hub) *WebsocketServices {
	return &WebsocketServices{
		Hub: h,
	}
}
