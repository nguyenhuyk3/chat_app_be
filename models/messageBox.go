package models

type Payload struct {
	Type    string `json:"type"`
	Content []byte `json:"content"`
}

type Message struct {
	RoomId     string  `json:"roomId"`
	SenderId   string  `json:"senderId"`
	ReceiverId string  `json:"receiverId"`
	Payload    Payload `json:"payload"`
	State      string  `json:"state"`
	CreatedAt  string  `json:"createdAt"`
}

type MessageBox struct {
	Messages  []Message `json:"messages"`
	CreatedAt string    `json:"createdAt"`
}
