package notification

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/messaging"
)

func (n *NotificationServices) isTokenDifferent(userToken models.UserToken) (bool, *firestore.DocumentRef, error) {
	ctx := context.Background()
	query := n.FireStoreClient.Collection("tokenDevices").Where("userId", "==", userToken.UserId).Limit(1)
	docs, err := query.Documents(ctx).GetAll()

	if err != nil {
		return false, nil, fmt.Errorf("(isTokenDifferent): %v", err)
	}
	if len(docs) > 0 {
		existingDoc := docs[0]
		existingToken := existingDoc.Data()["token"].(string)
		if existingToken != userToken.Token {
			// Other tokens, need to update
			return true, existingDoc.Ref, nil
		}
		return false, existingDoc.Ref, nil
	}
	// UserId not found, need to add new one
	return true, nil, nil
}

type MessageNotification struct {
	Token  string
	Avatar string
	Title  string
	Body   string
}

func SendNotificationForCommingMessage(messagingClient *messaging.Client, messageNotification MessageNotification) error {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title:    messageNotification.Title,
			Body:     messageNotification.Body,
			ImageURL: messageNotification.Avatar,
		},
		Token: messageNotification.Token,
	}

	fmt.Println(messageNotification.Token)

	response, err := messagingClient.Send(context.Background(), message)

	if err != nil {
		return fmt.Errorf("having error when sending notification (SendNotificationForCommingMessage): %v", err)
	}

	log.Printf("send notification for token (%s) sucessfully: %s\n", response, message.Token)

	return nil
}
