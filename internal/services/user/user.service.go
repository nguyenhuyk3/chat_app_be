package user

import (
	"be_chat_app/models"
	"context"
	"errors"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServices struct {
	FireStoreClient *firestore.Client
}

func NewUserServices(client *firestore.Client) *UserServices {
	return &UserServices{
		FireStoreClient: client,
	}
}

type SearchUserReponse struct {
	UserId string `json:"user_id"`
}

func (u *UserServices) SearchUserIdByEmail(email string) (SearchUserReponse, int, error) {
	if email == "" {
		return SearchUserReponse{}, http.StatusBadRequest, errors.New("email should not be empty")
	}

	iter := u.FireStoreClient.Collection("users").Where("email", "==", email).Documents(context.Background())

	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return SearchUserReponse{}, http.StatusInternalServerError, fmt.Errorf("error getting document: %v", err)
		}

		return SearchUserReponse{UserId: doc.Ref.ID}, http.StatusOK, nil
	}

	return SearchUserReponse{}, http.StatusNotFound, errors.New("user not found")
}

func (u *UserServices) CheckUserIfExistByEmail(email string) (bool, int, error) {
	iter := u.FireStoreClient.Collection("users").Where("email", "==", email).Documents(context.Background())
	_, err := iter.Next()

	if err != nil {
		if err == iterator.Done {
			return false, http.StatusNotFound, nil
		}

		return false, http.StatusInternalServerError, fmt.Errorf("error checking user: %v", err)
	}

	return true, http.StatusOK, nil
}

func (u *UserServices) GetUserByEmail(email string) (models.User, int, error) {
	docRef := u.FireStoreClient.Collection("users").Where("email", "==", email).Documents(context.Background())

	for {
		doc, err := docRef.Next()
		// Check if there is any document
		if err == iterator.Done {
			return models.User{}, http.StatusNotFound, fmt.Errorf("user with email %s not found", email)
		}
		// If having error in process gets document
		if err != nil {
			return models.User{}, http.StatusInternalServerError, fmt.Errorf("error fetching user: %v", err)
		}

		var user models.User

		err = doc.DataTo(&user)
		if err != nil {
			return models.User{}, http.StatusInternalServerError, fmt.Errorf("error mapping document data to user: %v", err)
		}

		return user, http.StatusOK, nil
	}
}

func (u *UserServices) GetUserById(userId string) (interface{}, int, error) {
	docRef := u.FireStoreClient.Collection("users").Doc(userId)
	docSnap, err := docRef.Get(context.Background())

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, http.StatusNotFound, fmt.Errorf("user with ID %s not found", userId)
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("error retrieving user document: %v", err)
	}

	var user models.User

	err = docSnap.DataTo(&user)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("error mapping document data to user struct: %v", err)
	}

	return user, http.StatusOK, nil
}

func (u *UserServices) GetSendingInvitationBox(invitationBoxId string) (models.SendingInvitation, int, error) {
	docRef := u.FireStoreClient.Collection("sendingInvitationBoxes").Doc(invitationBoxId)
	docSnap, err := docRef.Get(context.Background())

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.SendingInvitation{}, http.StatusNotFound, fmt.Errorf("not found")
		}

		return models.SendingInvitation{}, http.StatusInternalServerError, fmt.Errorf("error retrieving: %v", err)
	}

	var sendingInvitation models.SendingInvitation
	if err := docSnap.DataTo(&sendingInvitation); err != nil {
		return models.SendingInvitation{}, http.StatusInternalServerError, fmt.Errorf("error mapping document data to SendingInvitation struct: %v", err)
	}

	return sendingInvitation, http.StatusOK, nil
}

func (u *UserServices) GetReceivingInvitationBox(invitationBoxId string) (models.ReceivingInvitation, int, error) {
	docRef := u.FireStoreClient.Collection("receivingInvitationBoxes").Doc(invitationBoxId)
	docSnap, err := docRef.Get(context.Background())

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.ReceivingInvitation{}, http.StatusNotFound, fmt.Errorf("not found")
		}

		return models.ReceivingInvitation{}, http.StatusInternalServerError, fmt.Errorf("error retrieving: %v", err)
	}

	var receivingingInvitation models.ReceivingInvitation
	if err := docSnap.DataTo(&receivingingInvitation); err != nil {
		return models.ReceivingInvitation{}, http.StatusInternalServerError, fmt.Errorf("error mapping document data to SendingInvitation struct: %v", err)
	}

	return receivingingInvitation, http.StatusOK, nil
}
