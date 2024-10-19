package user

import (
	"be_chat_app/models"
	"fmt"
	"net/http"
	"sync"
)

func (u *UserServices) AddMessageBoxForBothUser(
	firstUserInfor, secondUserInfor models.InforUser) (string, int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var finalMessageBoxId string
	var finalErr error
	var finalStatus int

	wg.Add(1)
	go func() {
		defer wg.Done()

		var err error
		messageBoxId, _, err := u.createMessageBox(firstUserInfor, secondUserInfor)
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
		return "", finalStatus, fmt.Errorf("%v", finalErr)
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		status, err := u.addMessageBoxIntoUser(finalMessageBoxId, firstUserInfor.Email)
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

		status, err := u.addMessageBoxIntoUser(finalMessageBoxId, secondUserInfor.Email)
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
		return "", finalStatus, fmt.Errorf("%v", finalErr)
	}
	return finalMessageBoxId, http.StatusOK, nil
}

func (u *UserServices) ReadUnreadedMessages(messageBoxId, userId string) (int, error) {
	type result struct {
		status int
		err    error
	}

	// Buffer size of 2 to avoid blocking
	results := make(chan result, 2)
	var finalStatus int

	// Goroutine 1: Mark messages as read
	go func() {
		status, err := u.markReadedMessagesWhenJoiningMessageBox(messageBoxId, userId)
		results <- result{status: status, err: err}
	}()

	// Goroutine 2: Update last state for user
	go func() {
		status, err := u.updateLastStateForUser(messageBoxId, userId)
		results <- result{status: status, err: err}
	}()

	var finalError error

	for i := 0; i < 2; i++ {
		res := <-results
		if res.err != nil {
			finalError = res.err
			finalStatus = res.status
		}
	}

	if finalError != nil {
		return finalStatus, fmt.Errorf("%v", finalError)
	}
	return http.StatusOK, nil
}
