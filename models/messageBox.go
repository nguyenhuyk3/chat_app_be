package models

import "time"

type Notification struct {
	FromUserInfor
	ToUserId  string `json:"toUserId" firestore:"toUserId"`
	Content   string `json:"content" firestore:"content"`
	CreatedAt string `json:"createdAt" firestore:"createdAt"`
}

// type Payload struct {
// 	Type    string `json:"type" firestore:"type"`
// 	Content []byte `json:"content" firestore:"content"`
// }

type CommingMessage struct {
	MessageBoxId string    `json:"messageBoxId" firestore:"messageBoxId"`
	SenderId     string    `json:"senderId" firestore:"senderId"`
	TokenDevice  string    `json:"tokenDevice" firestore:"tokenDevice"`
	ReceiverId   string    `json:"receiverId" firestore:"receiverId"`
	Type         string    `json:"type" firestore:"type"`
	Content      string    `json:"content" firestore:"content"`
	SendedId     string    `json:"sendedId" firestore:"sendedId"`
	State        string    `json:"state" firestore:"state"`
	CreatedAt    time.Time `json:"createdAt" firestore:"createdAt"`
}

type Message struct {
	SenderId string `json:"senderId" firestore:"senderId"`
	// ReceiverId   string  `json:"receiverId" firestore:"receiverId"`
	Type      string `json:"type" firestore:"type"`
	Content   string `json:"content" firestore:"content"`
	SendedId  string `json:"sendedId" firestore:"sendedId"`
	State     string `json:"state" firestore:"state"`
	CreatedAt string `json:"createdAt" firestore:"createdAt"`
}

type InforUser struct {
	Id       string `json:"id" firestore:"id"`
	Email    string `json:"email" firestore:"email"`
	FullName string `json:"fullName" firestore:"fullName"`
	Avtar    string `json:"avatar" firestore:"avatar"`
}

type LastState struct {
	UserId      string `json:"userId" fireStore:"userId"`
	LastMessage string `json:"lastMessage" firestore:"lastMessage"`
	LastTime    string `json:"lastTime" firestore:"lastTime"`
	LastStatus  string `json:"lastStatus" firestore:"lastStatus"`
}

type MessageBox struct {
	FirstInforUser                InforUser `json:"firstInforUser" firestore:"firstInforUser"`
	SecondInforUser               InforUser `json:"secondInforUser" firestore:"secondInforUser"`
	LastStateMessageForFirstUser  LastState `json:"lastStateMessageForFirstUser" firestore:"lastStateMessageForFirstUser"`
	LastStateMessageForSecondUser LastState `json:"lastStateMessageForSecondUser" firestore:"lastStateMessageForSecondUser"`
	Messages                      []Message `json:"messages" firestore:"messages"`
	CreatedAt                     string    `json:"createdAt" firestore:"createdAt"`
}

type MessageBoxResponse struct {
	MessageBoxId string `json:"messageBoxId"`
	MessageBox
}

type ReadedNotification struct {
	MessageBoxId string `json:"messageBoxId"`
	UserId       string `json:"userId"`
	Readed       bool   `json:"readed"`
}
