package models

type Information struct {
	FullName   string `json:"fullName" firestore:"fullName"`
	Genre      string `json:"genre" firestore:"genre"`
	DayOfBirth string `json:"dayOfBirth" firestore:"dayOfBirth"`
}

type MessageBoxes []string

type FromUserInfor struct {
	FromUserEmail string `json:"fromUserEmail" firestore:"fromUserEmail"`
	FromUserName  string `json:"fromUserName" firestore:"fromUserName"`
}

type FriendRequest struct {
	// Id            string        `json:"id" firestore:"id"`
	FromUserInfor FromUserInfor `json:"fromUserInfor" firestore:"fromUserInfor"`
	ToUserEmail   string        `json:"toUserEmail" firestore:"toUserEmail"`
	Status        string        `json:"status" firestore:"status"`
	CreatedAt     string        `json:"createdAt" firestore:"createdAt"`
}

type SendingInvitation struct {
	// OwnerId        string          `json:"ownerId" firestore:"ownerId"`
	FriendRequests []FriendRequest `json:"friendRequests" firestore:"friendRequests"`
}

type ReceivingInvitation struct {
	// OwnerId        string          `json:"ownerId" firestore:"ownerId"`
	FriendRequests []FriendRequest `json:"friendRequests" firestore:"friendRequests"`
}

type User struct {
	Id                     string       `json:"id" firestore:"id"`
	PhoneNumber            string       `json:"phoneNumber" firestore:"phoneNumber"`
	Email                  string       `json:"email" firestore:"email"`
	HashPassword           string       `json:"hashPassword" firestore:"hashPassword"`
	SendingInvitationBox   string       `json:"sendingInvitationBox" firestore:"sendingInvitationBox"`
	ReceivingInvitationBox string       `json:"receivingInvitationBox" firestore:"receivingInvitationBox"`
	Friends                []string     `json:"friends" firestore:"friends"`
	Information            Information  `json:"information" firestore:"information"`
	MessageBoxes           MessageBoxes `json:"messageBoxes" firestore:"messageBoxes"`
	CreatedAt              string       `json:"createdAt" firestore:"createdAt"`
	UpdatedAt              string       `json:"updatedAt" firestore:"updatedAt"`
}
