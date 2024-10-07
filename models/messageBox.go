package models

type Notification struct {
	FromUserInfor
	ToUserId  string `json:"toUserId" firestore:"toUserId"`
	Content   string `json:"content" firestore:"content"`
	CreatedAt string `json:"createdAt" firestore:"createdAt"`
}

type Payload struct {
	Type    string `json:"type" firestore:"type"`
	Content []byte `json:"content" firestore:"content"`
}

type Message struct {
	MessageBoxId string  `json:"messageBoxId" firestore:"messageBoxId"`
	SenderId     string  `json:"senderId" firestore:"senderId"`
	ReceiverId   string  `json:"receiverId" firestore:"receiverId"`
	Payload      Payload `json:"payload" firestore:"payload"`
	State        string  `json:"state" firestore:"state"`
	CreatedAt    string  `json:"createdAt" firestore:"createdAt"`
}

type InforUser struct {
	Id       string `json:"id" firestore:"id"`
	Email    string `json:"email" firestore:"email"`
	FullName string `json:"fullName" firestore:"fullName"`
	Avtar    string `json:"avatar" firestore:"avatar"`
}

type LastState struct {
	LastMessage string `json:"lastMessage" firestore:"lastMessage"`
	LastTime    string `json:"lastTime" firestore:"lastTime"`
}

type MessageBox struct {
	FirstInforUser   InforUser `json:"firstInforUser" firestore:"firstInforUser"`
	SecondInforUser  InforUser `json:"secondInforUser" firestore:"secondInforUser"`
	LastStateMessage LastState `json:"lastStateMessage" firestore:"lastStateMessage"`
	Messages         []Message `json:"messages" firestore:"messages"`
	CreatedAt        string    `json:"createdAt" firestore:"createdAt"`
}
