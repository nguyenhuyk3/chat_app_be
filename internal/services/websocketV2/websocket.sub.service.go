package websocketv2

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func (w *WebsocketServices) saveBatchMessages(commingMessages []models.CommingMessage) (int, error) {
	batch := w.FireStoreClient.BulkWriter(context.Background())

	for _, commingMessage := range commingMessages {
		docRef := w.FireStoreClient.Collection("messageBoxes").Doc(commingMessage.MessageBoxId)
		docSnap, err := docRef.Get(context.Background())
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to get messageBoxes: %v", err)
		}

		var messageBox models.MessageBox
		if err := docSnap.DataTo(&messageBox); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to convert to messageBoxes: %v", err)
		}

		var stateForSender, stateForReceiver, stateForBoth string
		if len(w.Hub.MessageBoxes[commingMessage.MessageBoxId].Clients) > 1 {
			stateForSender = "đã đọc"
			stateForReceiver = "đã đọc"
			stateForBoth = "đã đọc"
		} else {
			stateForSender = "đã đọc"
			stateForReceiver = "chưa đọc"
			stateForBoth = "chưa đọc"
		}

		newMessage := models.Message{
			SenderId:  commingMessage.SenderId,
			Type:      commingMessage.Type,
			Content:   commingMessage.Content,
			SendedId:  commingMessage.SendedId,
			State:     stateForBoth,
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		messageBox.LastStateMessageForFirstUser = models.LastState{
			UserId:      commingMessage.SenderId,
			LastMessage: string(newMessage.Content),
			LastTime:    time.Now().Format("2006-01-02 15:04:05"),
			LastStatus:  stateForSender,
		}
		messageBox.LastStateMessageForSecondUser = models.LastState{
			UserId:      commingMessage.ReceiverId,
			LastMessage: string(newMessage.Content),
			LastTime:    time.Now().Format("2006-01-02 15:04:05"),
			LastStatus:  stateForReceiver,
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

func (w *WebsocketServices) MarkReadedMessagesWhenJoinMessageBox(messageBoxId, userId string) (int, error) {
	docRef := w.FireStoreClient.Collection("messageBoxes").Doc(messageBoxId).Collection("messages")
	query := docRef.Where("state", "==", "chưa đọc").Where("senderId", "!=", userId)
	iter := query.Documents(context.Background())

	defer iter.Stop()

	var batch = w.FireStoreClient.Batch()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("error retrieving messages (MarkReadedMessagesWhenJoinMessageBox): %v", err)
		}

		batch.Update(doc.Ref, []firestore.Update{
			{Path: "state", Value: "đã đọc"},
		})
	}

	_, err := batch.Commit(context.Background())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error updating messages: %v", err)
	}
	return http.StatusOK, nil
}

func (w *WebsocketServices) UpdateLastStateForUser(messageBoxId, userId, lastMessage, lastTime string) (int, error) {
	docRef := w.FireStoreClient.Collection("messageBoxes").Doc(messageBoxId)
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
		path = "lastStateMessageForFirstUser"
	} else {
		path = "lastStateMessageForSecondUser"
	}

	lastMessagePath := path + ".lastMessage"
	lastTimePath := path + ".lastTime"
	lastStatus := path + ".lastStatus"

	_, err = docRef.Update(context.Background(), []firestore.Update{
		{Path: lastMessagePath, Value: lastMessage},
		{Path: lastTimePath, Value: lastTime},
		{Path: lastStatus, Value: lastStatus},
	})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error updating last state: %v", err)
	}
	return http.StatusOK, nil
}
