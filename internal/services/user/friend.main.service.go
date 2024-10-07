package user

import (
	"be_chat_app/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (u *UserServices) AddFriendReqToBox(
	collectionName, userId, friendRequestBoxId string,
	makeFriendReq models.FriendRequest) (int, error) {
	makeFriendReq.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	if friendRequestBoxId == "" {
		invitationId, status, err := u.CreateNewInvitationDoc(collectionName, makeFriendReq)
		if err != nil {
			return status, fmt.Errorf("%v", err)
		}

		return u.UpdateUserInvitationBox(userId, collectionName, invitationId)
	}

	docRef := u.FireStoreClient.Collection(collectionName).Doc(friendRequestBoxId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			invitationId, status, err := u.CreateNewInvitationDocHavingDocRef(docRef, collectionName, userId, makeFriendReq)
			if err != nil {
				return status, fmt.Errorf("%v", err)
			}

			return u.UpdateUserInvitationBox(userId, collectionName, invitationId)
		}

		return http.StatusNotFound, fmt.Errorf("failed to retrieve document: %v", err)
	}

	return u.UpdateExistingInvitation(docRef, docSnap, collectionName, makeFriendReq)
}

// fromUserEmail is a sender who send invitation not person who perform add-friend action on UI
func (u *UserServices) AcceptFriend(toUserEmail, fromUserEmail string) (string, string, int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var finalStatus int
	var finalErr error
	var finalFromUserId, finalToUserId string

	wg.Add(4)
	go func() {
		defer wg.Done()

		sendingInvitationBoxId, status, err := u.FindSubIdInDoc("sendingInvitationBoxId", fromUserEmail)
		if err != nil {
			mu.Lock()
			finalStatus = status
			finalErr = err
			mu.Unlock()
			return
		}

		status, err = u.DeleteMakingFriendReq("sendingInvitationBoxes", sendingInvitationBoxId, fromUserEmail)
		if err != nil {
			mu.Lock()
			finalStatus = status
			finalErr = err
			mu.Unlock()
			return
		}
	}()
	go func() {
		defer wg.Done()

		receivingInvitationBoxId, status, err := u.FindSubIdInDoc("receivingInvitationBoxId", toUserEmail)
		if err != nil {
			mu.Lock()
			finalStatus = status
			finalErr = err
			mu.Unlock()
			return
		}

		status, err = u.DeleteMakingFriendReq("receivingInvitationBoxes", receivingInvitationBoxId, toUserEmail)
		if err != nil {
			mu.Lock()
			finalStatus = status
			finalErr = err
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()

		fromUserId, _, err := u.SearchUserIdByEmail(fromUserEmail)
		if err != nil {
			mu.Lock()
			finalStatus = http.StatusInternalServerError
			finalErr = err
			mu.Unlock()
			return
		}

		finalFromUserId = fromUserId.UserId

		status, err := u.AddFriend(toUserEmail, fromUserId.UserId)
		if err != nil {
			mu.Lock()
			finalStatus = status
			finalErr = err
			mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()

		toUserId, _, err := u.SearchUserIdByEmail(toUserEmail)
		if err != nil {
			mu.Lock()
			finalStatus = http.StatusInternalServerError
			finalErr = err
			mu.Unlock()
			return
		}

		finalToUserId = toUserId.UserId

		status, err := u.AddFriend(fromUserEmail, toUserId.UserId)
		if err != nil {
			mu.Lock()
			finalStatus = status
			finalErr = err
			mu.Unlock()
		}
	}()

	wg.Wait()

	if finalErr != nil {
		return "", "", finalStatus, finalErr
	}

	return finalFromUserId, finalToUserId, http.StatusOK, nil
}

func (u *UserServices) GetSubIdsByEmail(email string) (SubIds, int, error) {
	user, userId, status, err := u.GetUserByEmail(email)
	if err != nil {
		return SubIds{}, status, fmt.Errorf("%v", err)
	}

	return SubIds{
		UserId:                   userId,
		SendingInvitationBoxId:   user.SendingInvitationBoxId,
		ReceivingInvitationBoxId: user.ReceivingInvitationBoxId,
	}, http.StatusOK, nil
}

func (u *UserServices) DeleteFriendRequest(fromUserBoxType, toUserBoxType,
	invitationBoxIdOfOwner,
	fromUserEmail, toUserEmail string) (int, error) {
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(2)

	go func() {
		defer wg.Done()

		if _, err := u.DeleteFriendRequestFromUser(fromUserBoxType, invitationBoxIdOfOwner, fromUserEmail, toUserEmail); err != nil {
			errCh <- fmt.Errorf("%s friend from request error: %v", fromUserBoxType, err)
		}
	}()

	go func() {
		defer wg.Done()

		if _, err := u.DeleteFriendRequestToUser(toUserBoxType, toUserEmail, fromUserEmail); err != nil {
			errCh <- fmt.Errorf("%s friend to request error: %v", toUserBoxType, err)
		}
	}()

	wg.Wait()
	close(errCh)

	var finalErr error

	for err := range errCh {
		if err != nil {
			finalErr = err

			break
		}
	}

	if finalErr != nil {
		return http.StatusInternalServerError, finalErr
	}

	return http.StatusOK, nil
}

func (u *UserServices) GetMessageBoxesByUserId(userId string) ([]string, int, error) {
	userDoc, _, _ := u.GetUserById(userId)

	user, ok := userDoc.(models.User)
	if !ok {
		return nil, http.StatusInternalServerError, errors.New("failed to covert userDoc")
	}

	return user.MessageBoxes, http.StatusOK, nil
}
