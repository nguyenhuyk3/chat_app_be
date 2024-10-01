package user

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (u *UserServices) BuildNewInvitation(collectionName string, makeFriendReq models.FriendRequest) interface{} {
	if collectionName == "sendingInvitations" {
		return models.SendingInvitation{
			// OwnerId:        userId,
			FriendRequests: []models.FriendRequest{makeFriendReq},
		}
	}
	return models.ReceivingInvitation{
		// OwnerId:        userId,
		FriendRequests: []models.FriendRequest{makeFriendReq},
	}
}

func (u *UserServices) CreateNewInvitationDoc(collectionName string, makeFriendReq models.FriendRequest) (string, int, error) {
	// Build the new invitation
	invitation := u.BuildNewInvitation(collectionName, makeFriendReq)

	// Add the new invitation to Firestore with an auto-generated ID
	docRef, _, err := u.FireStoreClient.Collection(collectionName).Add(context.Background(), invitation)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to add new invitation to %s: %v", collectionName, err)
	}

	return docRef.ID, http.StatusOK, nil
}

func (u *UserServices) CreateNewInvitationDocHavingDocRef(docRef *firestore.DocumentRef, collectionName, userId string, makeFriendReq models.FriendRequest) (string, int, error) {
	invitation := u.BuildNewInvitation(collectionName, makeFriendReq)

	_, err := docRef.Set(context.Background(), invitation)
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to create new %s invitation: %v", collectionName, err)
	}

	return docRef.ID, http.StatusOK, nil
}

func (u *UserServices) UpdateExistingInvitation(docRef *firestore.DocumentRef, docSnap *firestore.DocumentSnapshot, collectionName string, makeFriendReq models.FriendRequest) (int, error) {
	var invitation interface{}

	if collectionName == "sendingInvitations" {
		var sendingInvitation models.SendingInvitation
		err := docSnap.DataTo(&sendingInvitation)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to map document data: %v", err)
		}

		sendingInvitation.FriendRequests = append(sendingInvitation.FriendRequests, makeFriendReq)
		invitation = sendingInvitation
	} else {
		var receivingInvitation models.ReceivingInvitation
		err := docSnap.DataTo(&receivingInvitation)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to map document data: %v", err)
		}

		receivingInvitation.FriendRequests = append(receivingInvitation.FriendRequests, makeFriendReq)
		invitation = receivingInvitation
	}

	_, err := docRef.Set(context.Background(), invitation)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update %s invitations: %v", collectionName, err)
	}

	return http.StatusOK, nil
}

func (u *UserServices) UpdateUserInvitationBox(userId, collectionName, invitationId string) (int, error) {
	userDoc, status, err := u.GetUserById(userId)
	if err != nil {
		return status, fmt.Errorf("%v", err)
	}

	user, ok := userDoc.(models.User)
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("failed to convert user document to User type")
	}

	if collectionName == "sendingInvitationBoxes" {
		user.SendingInvitationBox = invitationId
	} else if collectionName == "receivingInvitationBoxes" {
		user.ReceivingInvitationBox = invitationId
	}

	userDocRef := u.FireStoreClient.Collection("users").Doc(userId)
	_, err = userDocRef.Set(context.Background(), user)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user document: %v", err)
	}

	return http.StatusOK, nil
}

func (u *UserServices) AddFriendReqToBox(collectionName, userId, friendRequestBoxId string, makeFriendReq models.FriendRequest) (int, error) {
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

func (u *UserServices) FindSubIdInDoc(subIdName, email string) (string, int, error) {
	user, status, err := u.GetUserByEmail(email)

	fmt.Println(user)
	if err != nil {
		return "", status, fmt.Errorf("%v", err)
	}

	switch subIdName {
	case "receivingInvitationBox":
		return user.ReceivingInvitationBox, http.StatusOK, nil
	case "sendingInvitationBox":
		return user.SendingInvitationBox, http.StatusOK, nil
	default:
		return "", http.StatusBadRequest, fmt.Errorf("subId %s not found in %s struct", subIdName, email)
	}
}

func (u *UserServices) DeleteMakingFriendReq(invitationType, invitationBoxId, email string) (int, error) {
	docRef := u.FireStoreClient.Collection(invitationType).Doc(invitationBoxId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("%s with ID: %s not found", invitationType, invitationBoxId)
		}

		return http.StatusInternalServerError, fmt.Errorf("error retrieving %s document: %v", invitationType, err)
	}

	switch invitationType {
	case "receivingInvitationBoxes":
		var invitationBox models.ReceivingInvitation

		err = docSnap.DataTo(&invitationBox)

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("error mapping document data to %s struct: %v", invitationType, err)
		}

		var updatedFriendRequests []models.FriendRequest

		for _, friendReq := range invitationBox.FriendRequests {
			if friendReq.ToUserEmail != email {
				updatedFriendRequests = append(updatedFriendRequests, friendReq)
			}
		}

		if len(updatedFriendRequests) == len(invitationBox.FriendRequests) {
			return http.StatusNotFound, fmt.Errorf("no friend request found with email: %s", email)
		}

		invitationBox.FriendRequests = updatedFriendRequests

		_, err = docRef.Set(context.Background(), invitationBox)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to update sending invitation: %v", err)
		}

		return http.StatusOK, nil
	case "sendingInvitationBoxes":
		var invitationBox models.SendingInvitation

		err = docSnap.DataTo(&invitationBox)

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("error mapping document data to %s struct: %v", invitationType, err)
		}

		var updatedFriendRequests []models.FriendRequest

		for _, friendReq := range invitationBox.FriendRequests {
			if friendReq.FromUserInfor.FromUserEmail != email {
				updatedFriendRequests = append(updatedFriendRequests, friendReq)
			}
		}

		if len(updatedFriendRequests) == len(invitationBox.FriendRequests) {
			return http.StatusNotFound, fmt.Errorf("no friend request found with email: %s", email)
		}

		invitationBox.FriendRequests = updatedFriendRequests

		_, err = docRef.Set(context.Background(), invitationBox)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to update %s: %v", invitationType, err)
		}

		return http.StatusOK, nil
	default:
		return http.StatusBadRequest, fmt.Errorf("invalid invitation type: %s", invitationType)
	}
}

func (u *UserServices) AddFriend(ownerEmail, newFriendId string) (int, error) {
	userRef := u.FireStoreClient.Collection("users").Where("email", "==", ownerEmail)
	docSnap, err := userRef.Documents(context.Background()).Next()
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return http.StatusNotFound, fmt.Errorf("User with email: %s not found", ownerEmail)
		}

		return http.StatusInternalServerError, fmt.Errorf("error retrieving user by email: %v", err)
	}

	var user models.User

	err = docSnap.DataTo(&user)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("error mapping document data to User struct: %v", err)
	}

	for _, friend := range user.Friends {
		if friend == newFriendId {
			return http.StatusConflict, fmt.Errorf("Friend with id: %s already exists", newFriendId)
		}
	}

	user.Friends = append(user.Friends, newFriendId)

	_, err = docSnap.Ref.Set(context.Background(), user)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user: %v", err)
	}

	return http.StatusOK, nil
}

func (u *UserServices) AcceptFriend(toUserEmail, fromUserEmail string) (int, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var finalStatus int
	var finalErr error

	wg.Add(4)

	go func() {
		defer wg.Done()

		sendingInvitationBoxId, status, err := u.FindSubIdInDoc("sendingInvitationBox", fromUserEmail)
		fmt.Println(sendingInvitationBoxId)
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

		receivingInvitationBoxId, status, err := u.FindSubIdInDoc("receivingInvitationBox", toUserEmail)
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
		return finalStatus, finalErr
	}

	return http.StatusOK, nil
}
