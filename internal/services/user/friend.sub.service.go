package user

import (
	"be_chat_app/models"
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (u *UserServices) BuildNewInvitation(collectionName string, makeFriendReq models.FriendRequest) interface{} {
	if collectionName == "sendingInvitations" {
		return models.SendingInvitationBox{
			// OwnerId:        userId,
			FriendRequests: []models.FriendRequest{makeFriendReq},
		}
	}
	return models.ReceivingInvitationBox{
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
		var sendingInvitation models.SendingInvitationBox
		err := docSnap.DataTo(&sendingInvitation)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to map document data: %v", err)
		}

		sendingInvitation.FriendRequests = append(sendingInvitation.FriendRequests, makeFriendReq)
		invitation = sendingInvitation
	} else {
		var receivingInvitation models.ReceivingInvitationBox
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

func (u *UserServices) UpdateUserInvitationBox(userId, collectionName, invitationBoxId string) (int, error) {
	userDoc, status, err := u.GetUserById(userId)
	if err != nil {
		return status, fmt.Errorf("%v", err)
	}

	user, ok := userDoc.(models.User)
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("failed to convert user document to User type")
	}

	if collectionName == "sendingInvitationBoxes" {
		user.SendingInvitationBoxId = invitationBoxId
	} else if collectionName == "receivingInvitationBoxes" {
		user.ReceivingInvitationBoxId = invitationBoxId
	}

	userDocRef := u.FireStoreClient.Collection("users").Doc(userId)
	_, err = userDocRef.Set(context.Background(), user)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user document: %v", err)
	}
	return http.StatusOK, nil
}

func (u *UserServices) FindSubIdInDoc(subIdName, email string) (string, int, error) {
	user, _, status, err := u.GetUserByEmail(email)

	if err != nil {
		return "", status, fmt.Errorf("%v", err)
	}

	switch subIdName {
	case "receivingInvitationBoxId":
		return user.ReceivingInvitationBoxId, http.StatusOK, nil
	case "sendingInvitationBoxId":
		return user.SendingInvitationBoxId, http.StatusOK, nil
	default:
		return "", http.StatusBadRequest, fmt.Errorf("subId %s not found in %s struct", subIdName, email)
	}
}

type SubIds struct {
	UserId                   string `json:"userId"`
	ReceivingInvitationBoxId string `json:"receivingInvitationBoxId"`
	SendingInvitationBoxId   string `json:"sendingInvitationBoxId"`
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
		var invitationBox models.ReceivingInvitationBox

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
			return http.StatusNotFound, fmt.Errorf("(reiceiving) no friend request found with email: %s", email)
		}

		invitationBox.FriendRequests = updatedFriendRequests

		_, err = docRef.Set(context.Background(), invitationBox)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to update sending invitation: %v", err)
		}
		return http.StatusOK, nil
	case "sendingInvitationBoxes":
		var invitationBox models.SendingInvitationBox

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
			return http.StatusNotFound, fmt.Errorf("(sending) no friend request found with email: %s", email)
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
			return http.StatusNotFound, fmt.Errorf("user with email: %s not found", ownerEmail)
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
			return http.StatusConflict, fmt.Errorf("friend with id: %s already exists", newFriendId)
		}
	}

	user.Friends = append(user.Friends, newFriendId)

	_, err = docSnap.Ref.Set(context.Background(), user)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user: %v", err)
	}
	return http.StatusOK, nil
}

func (u *UserServices) DeleteFriendRequestFromUser(boxType, invitationBoxId, fromUserEmail, toUserEmail string) (int, error) {
	switch boxType {
	case "sendingInvitationBoxes":
		docRef := u.FireStoreClient.Collection(boxType).Doc(invitationBoxId)
		docSnap, err := docRef.Get(context.Background())

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("error retrieving sending document: %v", err)
		}

		if !docSnap.Exists() {
			return http.StatusNotFound, fmt.Errorf("sending document does not exist")
		}

		var invitations models.SendingInvitationBox

		if err := docSnap.DataTo(&invitations); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(from user) error mapping document data to %s box struct: %v", boxType, err)
		}

		var newInvitations []models.FriendRequest

		for _, v := range invitations.FriendRequests {
			if v.ToUserEmail != toUserEmail {
				newInvitations = append(newInvitations, v)
			}
		}

		if len(newInvitations) == len(invitations.FriendRequests) {
			return http.StatusInternalServerError, fmt.Errorf("(from user) FROM length should not be equal")
		}

		_, err = docRef.Update(context.Background(), []firestore.Update{
			{
				Path:  "friendRequests",
				Value: newInvitations,
			},
		})

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(from user) error updating sending invitation box: %v", err)
		}

		return http.StatusOK, nil
	case "receivingInvitationBoxes":
		docRef := u.FireStoreClient.Collection(boxType).Doc(invitationBoxId)
		docSnap, err := docRef.Get(context.Background())

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("error retrieving receiving document: %v", err)
		}

		if !docSnap.Exists() {
			return http.StatusNotFound, fmt.Errorf("receiving document does not exist")
		}

		var invitations models.ReceivingInvitationBox

		if err := docSnap.DataTo(&invitations); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(from user) error mapping document data to %s box struct: %v", boxType, err)
		}

		var newInvitations []models.FriendRequest

		for _, v := range invitations.FriendRequests {
			if v.FromUserInfor.FromUserEmail != toUserEmail {
				newInvitations = append(newInvitations, v)
			}
		}

		if len(newInvitations) == len(invitations.FriendRequests) {
			return http.StatusInternalServerError, fmt.Errorf("(to user) FROM length should not be equal")
		}

		_, err = docRef.Update(context.Background(), []firestore.Update{
			{
				Path:  "friendRequests",
				Value: newInvitations,
			},
		})

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(from user) error updating sending invitation box: %v", err)
		}
		return http.StatusOK, nil
	default:
		return http.StatusBadRequest, fmt.Errorf("(from user) invalid box type: %s", boxType)
	}
}

func (u *UserServices) DeleteFriendRequestToUser(boxType, toUserEmail, fromUserEmail string) (int, error) {
	switch boxType {
	case "sendingInvitationBoxes":
		subIds, _, _ := u.GetSubIdsByEmail(toUserEmail)
		docRef := u.FireStoreClient.Collection(boxType).Doc(subIds.SendingInvitationBoxId)
		docSnap, err := docRef.Get(context.Background())

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("error retrieving sending document: %v", err)
		}

		if !docSnap.Exists() {
			return http.StatusNotFound, fmt.Errorf("sending document does not exist")
		}

		var invitations models.SendingInvitationBox

		if err := docSnap.DataTo(&invitations); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(to user) error mapping document data to %s box struct: %v", boxType, err)
		}

		var newInvitations []models.FriendRequest

		for _, v := range invitations.FriendRequests {
			if v.ToUserEmail != fromUserEmail {
				newInvitations = append(newInvitations, v)
			}
		}

		if len(newInvitations) == len(invitations.FriendRequests) {
			return http.StatusInternalServerError, fmt.Errorf("(from user) TO length should not be equal")
		}

		_, err = docRef.Update(context.Background(), []firestore.Update{
			{
				Path:  "friendRequests",
				Value: newInvitations,
			},
		})

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(to user) error updating sending invitation box: %v", err)
		}

		return http.StatusOK, nil
	case "receivingInvitationBoxes":
		subIds, _, _ := u.GetSubIdsByEmail(toUserEmail)
		docRef := u.FireStoreClient.Collection(boxType).Doc(subIds.ReceivingInvitationBoxId)
		docSnap, err := docRef.Get(context.Background())

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("error retrieving receiving document: %v", err)
		}

		if !docSnap.Exists() {
			return http.StatusNotFound, fmt.Errorf("receiving document does not exist")
		}

		var invitations models.ReceivingInvitationBox

		if err := docSnap.DataTo(&invitations); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(to user) error mapping document data to %s box struct: %v", boxType, err)
		}

		var newInvitations []models.FriendRequest

		for _, v := range invitations.FriendRequests {
			if v.FromUserInfor.FromUserEmail != fromUserEmail {
				newInvitations = append(newInvitations, v)
			}
		}

		if len(newInvitations) == len(invitations.FriendRequests) {
			return http.StatusInternalServerError, fmt.Errorf("(to user) TO length should not be equal")
		}

		_, err = docRef.Update(context.Background(), []firestore.Update{
			{
				Path:  "friendRequests",
				Value: newInvitations,
			},
		})

		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("(to user) error updating sending invitation box: %v", err)
		}
		return http.StatusOK, nil
	default:
		return http.StatusBadRequest, fmt.Errorf("invalid box type: %s", boxType)
	}
}
