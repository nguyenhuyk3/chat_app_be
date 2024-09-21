package models

type Information struct {
	FullName   string
	Genre      string
	DayOfBirth string
}

type MessageBoxes []string

type User struct {
	Id           string
	PhoneNumber  string
	Email        string
	HashPassword string
	Information  Information
	MessageBoxes MessageBoxes
}
