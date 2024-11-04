package notification

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/iterator"
)

type NotificationServices struct {
	FireStoreClient *firestore.Client
	MessagingClient *messaging.Client
}

func NewNotificatioinServices(fireStoreClient *firestore.Client, messagingClient *messaging.Client) *NotificationServices {
	return &NotificationServices{
		FireStoreClient: fireStoreClient,
		MessagingClient: messagingClient,
	}
}

func (n *NotificationServices) SaveToken(userToken models.UserToken) (int, error) {
	shouldUpdate, docRef, err := n.isTokenDifferent(userToken)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	ctx := context.Background()

	if docRef != nil && shouldUpdate {
		// Update token if needed
		_, err = docRef.Set(ctx, userToken)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(UpdateToken): %v", err)
		}
		return http.StatusOK, nil
	}
	if docRef == nil {
		// Add new document if userId does not exist
		_, _, err = n.FireStoreClient.Collection("tokenDevices").Add(ctx, userToken)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(AddToken): %v", err)
		}
		return http.StatusOK, nil
	}
	return http.StatusNotModified, nil
}

func (n *NotificationServices) DeleteToken(userId string) (int, error) {
	docRef := n.FireStoreClient.Collection("tokenDevices").Where("userId", "==", userId).Documents(context.Background())
	tokenDeviceDoc, err := docRef.Next()
	if err != nil {
		if err == iterator.Done {
			return http.StatusNotFound, nil
		}
		return http.StatusInternalServerError, err
	}
	_, err = tokenDeviceDoc.Ref.Delete(context.Background())
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (n *NotificationServices) GetTokenByUserId(userId string) (string, int, error) {
	query := n.FireStoreClient.Collection("tokenDevices").Where("userId", "==", userId).Limit(1)
	docs, err := query.Documents(context.Background()).GetAll()
	if err != nil {
		return "", http.StatusOK, fmt.Errorf("(GetTokenByUserId): %v", err)
	}

	if len(docs) > 0 {
		doc := docs[0]
		token, ok := doc.Data()["token"].(string)
		if !ok {
			return "", http.StatusInternalServerError, fmt.Errorf("(GetTokenByUserId): token conversion error")
		}
		return token, http.StatusOK, nil
	}
	return "", http.StatusNotFound, fmt.Errorf("(GetTokenByUserId): no token found for userId %s", userId)
}
