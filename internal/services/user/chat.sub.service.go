package user

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

func (u *UserServices) CreateMessageBox(firstInfor, secondInfor models.InforUser) (string, int, error) {
	messageBox := models.MessageBox{
		FirstInforUser:  firstInfor,
		SecondInforUser: secondInfor,
		LastStateMessage: models.LastState{
			LastMessage: "Từ giờ hai bạn có thể trò chuyện",
			LastTime:    time.Now().Format("2006-01-02 15:04:05"),
		},
		Messages:  []models.Message{},
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	docRef, _, err := u.FireStoreClient.Collection("messageBoxes").Add(context.Background(), messageBox)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to add new messageBox: %v", err)
	}

	return docRef.ID, http.StatusOK, nil
}

func (u *UserServices) AddMessageBoxIntoUser(messageBoxId, email string) (int, error) {
	user, uesrId, _, _ := u.GetUserByEmail(email)

	user.MessageBoxes = append(user.MessageBoxes, messageBoxId)

	_, err := u.FireStoreClient.Collection("users").Doc(uesrId).Set(context.Background(), map[string]interface{}{
		"messageBoxes": user.MessageBoxes,
	}, firestore.MergeAll)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update %s user's messageBoxes: %v", email, err)
	}

	return http.StatusOK, nil
}
