package user

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"
	"sync"

	"cloud.google.com/go/firestore"
)

func (u *UserServices) ChangeInfomationAtRoot(userId string, newInformation models.Information) (int, error) {
	docRef := u.FireStoreClient.Collection("users").Doc(userId)

	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "information", Value: newInformation},
	})

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user information: %v", err)
	}
	return http.StatusOK, nil
}

func (u *UserServices) UpdateInformationAtMessageBox(messageBoxId, path string, newInformation models.InforUser) (int, error) {
	docRef := u.FireStoreClient.Collection("messageBoxes").Doc(messageBoxId)

	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: path, Value: newInformation},
	})

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user information (update information at message box): %v", err)
	}
	return http.StatusOK, nil
}

func (u *UserServices) ChangeInformationAtMessageBoxes(userId string, information models.Information) (int, error) {
	messageBoxesInterface, _, _ := u.GetAllMessageBoxesByUserId(userId)
	messageBoxes, ok := messageBoxesInterface.([]models.MessageBoxResponse)
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("cannot assert messageBoxResponse from returned data:")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var finalStatus int
	var finalErr error

	for _, messageBox := range messageBoxes {
		wg.Add(1)

		var newInforUser models.InforUser
		var isFirst bool

		if messageBox.FirstInforUser.Id == userId {
			newInforUser = models.InforUser{
				Id:       messageBox.FirstInforUser.Id,
				Email:    messageBox.FirstInforUser.Email,
				FullName: information.FullName,
				Avtar:    "",
			}
			isFirst = true
		} else {
			newInforUser = models.InforUser{
				Id:       messageBox.SecondInforUser.Id,
				Email:    messageBox.SecondInforUser.Email,
				FullName: information.FullName,
				Avtar:    "",
			}
			isFirst = false
		}

		go func(box models.MessageBoxResponse, user models.InforUser, first bool) {
			defer wg.Done()

			var path string

			if first {
				path = "firstInforUser"
			} else {
				path = "secondInforUser"
			}

			status, err := u.UpdateInformationAtMessageBox(box.MessageBoxId, path, newInforUser)
			if err != nil {
				mu.Lock()
				finalStatus = status
				finalErr = err
				mu.Unlock()
			}
		}(messageBox, newInforUser, isFirst)
	}

	wg.Wait()

	if finalErr != nil {
		return finalStatus, finalErr
	}
	return http.StatusOK, nil
}
