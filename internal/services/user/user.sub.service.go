package user

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
)

func (u *UserServices) changeInfomationAtRoot(userId string, newInformation models.Information) (int, error) {
	docRef := u.FireStoreClient.Collection("users").Doc(userId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "information", Value: newInformation},
	})

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user information: %v", err)
	}
	return http.StatusOK, nil
}

func (u *UserServices) updateInformationAtMessageBox(messageBoxId, path string, newInformation models.InforUser) (int, error) {
	docRef := u.FireStoreClient.Collection("messageBoxes").Doc(messageBoxId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: path, Value: newInformation},
	})

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user information (update information at message box): %v", err)
	}
	return http.StatusOK, nil
}
