package models

type FileMessage struct {
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	FileName   string `json:"fileName"`
	FileType   string `json:"fileType"`
	Content    []byte `json:"content"`
	CreatedAt  string `json:"createdAt"`
}
