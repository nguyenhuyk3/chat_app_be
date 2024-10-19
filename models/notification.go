package models

type UserToken struct {
	UserId string `json:"userId" firestore:"userId"`
	Token  string `json:"token" firestore:"token"`
}
