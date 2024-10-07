package user

import (
	"be_chat_app/models"
	"fmt"
	"net/http"
	"sync"
)

func (u *UserServices) AddMessageBoxForBothUser(
	firstUserInfor, secondUserInfor models.InforUser) (int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var finalMessageBoxId string
	var finalErr error
	var finalStatus int

	wg.Add(1)
	go func() {
		defer wg.Done()

		var err error
		messageBoxId, _, err := u.CreateMessageBox(firstUserInfor, secondUserInfor)
		if err != nil {
			mu.Lock()
			finalErr = err
			finalStatus = http.StatusInternalServerError
			mu.Unlock()
		}

		finalMessageBoxId = messageBoxId
	}()
	wg.Wait()

	if finalErr != nil {
		return finalStatus, fmt.Errorf("%v", finalErr)
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		status, err := u.AddMessageBoxIntoUser(finalMessageBoxId, firstUserInfor.Email)
		if err != nil {
			mu.Lock()
			finalErr = err
			finalStatus = status
			mu.Unlock()
			return
		}
	}()
	go func() {
		defer wg.Done()

		status, err := u.AddMessageBoxIntoUser(finalMessageBoxId, secondUserInfor.Email)
		if err != nil {
			mu.Lock()
			finalErr = err
			finalStatus = status
			mu.Unlock()
			return
		}
	}()
	wg.Wait()

	if finalErr != nil {
		return finalStatus, fmt.Errorf("%v", finalErr)
	}

	return http.StatusOK, nil
}
