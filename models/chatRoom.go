package models

type Members []string

type ChatRoom struct {
	MessageBoxId string  `json:"messageBoxId"`
	IsGroup      bool    `json:"isGroup"`
	Members      Members `json:"members"`
	CreatedBy    string  `json:"createdBy"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}
