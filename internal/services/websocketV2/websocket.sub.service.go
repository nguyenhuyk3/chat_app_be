package websocketv2

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

func (h *WebsocketServices) SaveBatchMessages(commingMessages []models.CommingMessage) (int, error) {
	batch := h.FireStoreClient.BulkWriter(context.Background())

	for _, commingMessage := range commingMessages {
		docRef := h.FireStoreClient.Collection("messageBoxes").Doc(commingMessage.MessageBoxId)
		docSnap, err := docRef.Get(context.Background())
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to get messageBoxes: %v", err)
		}

		var messageBox models.MessageBox
		if err := docSnap.DataTo(&messageBox); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to convert to messageBoxes: %v", err)
		}

		var state string
		if len(h.Hub.MessageBoxes[commingMessage.MessageBoxId].Clients) > 1 {
			state = "đã đọc"
		} else {
			state = "chưa đọc"
		}

		newMessage := models.Message{
			SenderId:  commingMessage.SenderId,
			Content:   commingMessage.Content,
			State:     state,
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		messageBox.LastStateMessageForFirstUser = models.LastState{
			UserId:      commingMessage.SenderId,
			LastMessage: string(newMessage.Content),
			LastTime:    time.Now().Format("2006-01-02 15:04:05"),
			LastStatus:  state,
		}
		messageBox.LastStateMessageForSecondUser = models.LastState{
			UserId:      commingMessage.ReceiverId,
			LastMessage: string(newMessage.Content),
			LastTime:    time.Now().Format("2006-01-02 15:04:05"),
			LastStatus:  state,
		}

		messageBox.Messages = append(messageBox.Messages, newMessage)

		if _, err := batch.Update(docRef, []firestore.Update{
			{Path: "messages", Value: messageBox.Messages},
			{Path: "lastStateMessageForFirstUser", Value: messageBox.LastStateMessageForFirstUser},
			{Path: "lastStateMessageForSecondUser", Value: messageBox.LastStateMessageForSecondUser},
		}); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to queue batch write: %v", err)
		}
	}

	batch.Flush()

	return http.StatusOK, nil
}
