package user

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

func (u *UserServices) createMessageBox(firstInfor, secondInfor models.InforUser) (string, int, error) {
	messageBox := models.MessageBox{
		FirstInforUser:  firstInfor,
		SecondInforUser: secondInfor,
		LastStateMessageForFirstUser: models.LastState{
			UserId:      firstInfor.Id,
			LastMessage: "Từ giờ hai bạn có thể trò chuyện",
			LastTime:    time.Now().Format("2006-01-02 15:04:05"),
			LastStatus:  "chưa đọc",
		},
		LastStateMessageForSecondUser: models.LastState{
			UserId:      secondInfor.Id,
			LastMessage: "Từ giờ hai bạn có thể trò chuyện",
			LastTime:    time.Now().Format("2006-01-02 15:04:05"),
			LastStatus:  "chưa đọc",
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

func (u *UserServices) addMessageBoxIntoUser(messageBoxId, email string) (int, error) {
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

func (u *UserServices) markReadedMessagesWhenJoiningMessageBox(messageBoxId, userId string) (int, error) {
	docRef := u.FireStoreClient.Collection("messageBoxes").Doc(messageBoxId)
	err := u.FireStoreClient.RunTransaction(context.Background(), func(ctx context.Context, tx *firestore.Transaction) error {
		docSnap, err := tx.Get(docRef)
		if err != nil {
			return fmt.Errorf("failed to retrieve messageBox: %v", err)
		}

		var messageBox models.MessageBox
		if err := docSnap.DataTo(&messageBox); err != nil {
			return fmt.Errorf("failed to convert messageBox: %v", err)
		}

		updated := false

		for i := len(messageBox.Messages) - 1; i >= 0; i-- {
			msg := &messageBox.Messages[i]

			if msg.State == "đã đọc" {
				break
			}
			if msg.State == "chưa đọc" && msg.SenderId != userId {
				msg.State = "đã đọc"
				updated = true
			}
		}

		if !updated {
			return nil
		}

		return tx.Update(docRef, []firestore.Update{
			{
				Path:  "messages",
				Value: messageBox.Messages,
			},
		})
	})

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update messages: %v", err)
	}
	return http.StatusOK, nil
}

func (u *UserServices) updateLastStateForUser(messageBoxId, userId string) (int, error) {
	docRef := u.FireStoreClient.Collection("messageBoxes").Doc(messageBoxId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get messageBoxes (UpdateLastStateForUser): %v", err)
	}

	var messageBox models.MessageBox
	if err := docSnap.DataTo(&messageBox); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to convert to messageBoxes (UpdateLastStateForUser): %v", err)
	}

	var path string
	if userId == messageBox.LastStateMessageForFirstUser.UserId {
		if messageBox.LastStateMessageForFirstUser.LastStatus == "đã đọc" {
			return http.StatusOK, nil
		}
		path = "lastStateMessageForFirstUser"
	} else if userId == messageBox.LastStateMessageForSecondUser.UserId {
		if messageBox.LastStateMessageForSecondUser.LastStatus == "đã đọc" {
			return http.StatusOK, nil
		}
		path = "lastStateMessageForSecondUser"
	}

	_, err = docRef.Update(context.Background(), []firestore.Update{
		{Path: fmt.Sprintf("%s.lastStatus", path), Value: "đã đọc"},
	})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error updating last state: %v", err)
	}
	return http.StatusOK, nil
}
