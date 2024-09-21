package models

import "time"

type Members []string

type ChatRoom struct {
	Id      string `json:"_id"`
	IsGroup bool   `json:"is_group"`
	Members Members
	// CreatedBy is user who create this room
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}
