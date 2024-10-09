package websocketv2

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"
	"time"
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

		newMessage := models.Message{
			SenderId:  commingMessage.SenderId,
			Content:   commingMessage.Content,
			State:     "chưa đọc",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		messageBox.LastStateMessage = models.LastState{
			LastMessage: string(newMessage.Content),
			LastTime:    time.Now().Format("2006-01-02 15:04:05"),
		}

		messageBox.Messages = append(messageBox.Messages, newMessage)

		if _, err := batch.Set(docRef, messageBox); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to queue batch write: %v", err)
		}
	}

	batch.Flush()

	return http.StatusOK, nil
}
