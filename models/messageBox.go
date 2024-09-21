package models

import "time"

type Payload struct {
	Type    string
	Content []byte
}

type Message struct {
	RoomId     string
	SenderId   string
	ReceiverId string
	Payload    Payload
	State      string
	CreatedAt  time.Time
}

type MessageBox struct {
	Messages  []Message
	CreatedAt time.Time
}
