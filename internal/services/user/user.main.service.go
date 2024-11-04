package user

import (
	"be_chat_app/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

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
	UserId string `json:"userId"`
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

func (u *UserServices) GetUserByEmail(email string) (models.User, string, int, error) {
	docRef := u.FireStoreClient.Collection("users").Where("email", "==", email).Documents(context.Background())

	for {
		doc, err := docRef.Next()
		// Check if there is any document
		if err == iterator.Done {
			return models.User{}, "", http.StatusNotFound, fmt.Errorf("user with email %s not found", email)
		}
		// If having error in process gets document
		if err != nil {
			return models.User{}, "", http.StatusInternalServerError, fmt.Errorf("error fetching user: %v", err)
		}

		var user models.User

		err = doc.DataTo(&user)
		if err != nil {
			return models.User{}, "", http.StatusInternalServerError, fmt.Errorf("error mapping document data to user: %v", err)
		}
		return user, doc.Ref.ID, http.StatusOK, nil
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

func (u *UserServices) GetUserIdByEmail(userEmail string) (string, int, error) {
	docRef := u.FireStoreClient.Collection("users").Where("email", "==", userEmail).Documents(context.Background())
	userDoc, err := docRef.Next()
	if err != nil {
		if err == iterator.Done {
			return "", 404, fmt.Errorf("user with email %s not found", userEmail)
		}
		return "", 500, fmt.Errorf("error querying Firestore: %v", err)
	}
	return userDoc.Ref.ID, http.StatusOK, nil
}

func (u *UserServices) GetSendingInvitationBox(invitationBoxId string) (models.SendingInvitationBox, int, error) {
	if invitationBoxId == "" {
		return models.SendingInvitationBox{}, http.StatusNotFound, fmt.Errorf("user have not sending invitation box id")
	}

	docRef := u.FireStoreClient.Collection("sendingInvitationBoxes").Doc(invitationBoxId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.SendingInvitationBox{}, http.StatusNotFound, fmt.Errorf("not found")
		}
		return models.SendingInvitationBox{}, http.StatusInternalServerError, fmt.Errorf("error retrieving: %v", err)
	}

	var sendingInvitation models.SendingInvitationBox

	if err := docSnap.DataTo(&sendingInvitation); err != nil {
		return models.SendingInvitationBox{}, http.StatusInternalServerError, fmt.Errorf("error mapping document data to SendingInvitation struct: %v", err)
	}
	return sendingInvitation, http.StatusOK, nil
}

func (u *UserServices) GetReceivingInvitationBox(invitationBoxId string) (models.ReceivingInvitationBox, int, error) {
	docRef := u.FireStoreClient.Collection("receivingInvitationBoxes").Doc(invitationBoxId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return models.ReceivingInvitationBox{}, http.StatusNotFound, fmt.Errorf("not found")
		}
		return models.ReceivingInvitationBox{}, http.StatusInternalServerError, fmt.Errorf("error retrieving: %v", err)
	}

	var receivingingInvitation models.ReceivingInvitationBox
	if err := docSnap.DataTo(&receivingingInvitation); err != nil {
		return models.ReceivingInvitationBox{}, http.StatusInternalServerError, fmt.Errorf("error mapping document data to SendingInvitation struct: %v", err)
	}
	return receivingingInvitation, http.StatusOK, nil
}

func (u *UserServices) GetMessageBoxById(messageBoxId string) (interface{}, int, error) {
	docRef := u.FireStoreClient.Collection("messageBoxes").Doc(messageBoxId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, http.StatusNotFound, fmt.Errorf("messageBox with ID %s not found", messageBoxId)
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("error retrieving user document: %v", err)
	}
	return docSnap.Data(), http.StatusOK, nil
}

func (u *UserServices) ChangeInformationAtRoot(userId string, newInformation models.Information) (int, error) {
	docRef := u.FireStoreClient.Collection("users").Doc(userId)

	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "information", Value: newInformation},
	})

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to update user information: %v", err)
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

			status, err := u.updateInformationAtMessageBox(box.MessageBoxId, path, newInforUser)
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

func (u *UserServices) GetFullNameById(userId string) (string, int, error) {
	docRef := u.FireStoreClient.Collection("users").Doc(userId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to get document: %v", err)
	}
	if info, ok := docSnap.Data()["information"].(map[string]interface{}); ok {
		if fullName, ok := info["fullName"].(string); ok {
			return fullName, http.StatusOK, nil
		} else {
			return "", http.StatusNotFound, fmt.Errorf("fullName field not found or not a string")
		}
	} else {
		return "", http.StatusNotFound, fmt.Errorf("information field not found or not a map")
	}
}

func (u *UserServices) GetFriendIdsById(userId string) []string {
	docRef := u.FireStoreClient.Collection("users").Doc(userId)
	docSnap, _ := docRef.Get(context.Background())

	var user models.User
	_ = docSnap.DataTo(&user)

	return user.Friends
}

func (u *UserServices) GetFriendEmailsById(userId string) ([]string, int, error) {
	friendIds := u.GetFriendIdsById(userId)
	friendEmails := []string{}

	for _, v := range friendIds {
		userEmail := u.getUserEmailByUserId(v)
		friendEmails = append(friendEmails, userEmail)
	}
	return friendEmails, http.StatusOK, nil
}

type InformationResponse struct {
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

func (u *UserServices) GetInformationByEmail(email string) (InformationResponse, int, error) {
	docRef := u.FireStoreClient.Collection("users").Where("email", "==", email).Documents(context.Background())

	for {
		doc, err := docRef.Next()
		if err == iterator.Done {
			return InformationResponse{}, http.StatusNotFound, fmt.Errorf("user with email %s not found (GetInformationByUserId)", email)
		}
		if err != nil {
			return InformationResponse{}, http.StatusInternalServerError, fmt.Errorf("error fetching user (GetInformationByUserId): %v", err)
		}

		var user models.User

		err = doc.DataTo(&user)
		if err != nil {
			return InformationResponse{}, http.StatusInternalServerError, fmt.Errorf("error mapping document data to user: %v", err)
		}
		return InformationResponse{Email: email, FullName: user.Information.FullName}, http.StatusOK, nil
	}
}

func (u *UserServices) GetEmailsFromInvitationBox(invitationBoxId, invitationType string) ([]string, int, error) {
	var emails []string
	var status int

	switch invitationType {
	case "receivingInvitationBoxes":
		docRef := u.FireStoreClient.Collection(invitationType).Doc(invitationBoxId)
		docSnap, err := docRef.Get(context.Background())
		if err != nil {
			return nil, 500, fmt.Errorf("failed to get document: %v", err) // HTTP 500 Internal Server Error
		}

		data := docSnap.Data()
		friendRequestsRaw, exists := data["friendRequests"]

		// If friendRequests is missing or null, return an empty list with a 200 status
		if !exists || friendRequestsRaw == nil {
			return emails, 200, nil
		}

		friendRequests, ok := friendRequestsRaw.([]interface{})
		if !ok {
			return nil, 400, fmt.Errorf("failed to parse friendRequests") // HTTP 400 Bad Request
		}

		for _, request := range friendRequests {
			requestMap, ok := request.(map[string]interface{})
			if !ok {
				continue
			}
			fromUserInfor, ok := requestMap["fromUserInfor"].(map[string]interface{})
			if !ok {
				continue
			}
			fromUserEmail, ok := fromUserInfor["fromUserEmail"].(string)
			if ok {
				emails = append(emails, fromUserEmail)
			}
		}
		status = 200 // HTTP 200 OK
		return emails, status, nil

	case "sendingInvitationBoxes":
		docRef := u.FireStoreClient.Collection(invitationType).Doc(invitationBoxId)
		docSnap, err := docRef.Get(context.Background())
		if err != nil {
			return nil, 500, fmt.Errorf("failed to get document: %v", err) // HTTP 500 Internal Server Error
		}

		data := docSnap.Data()
		friendRequestsRaw, exists := data["friendRequests"]

		// If friendRequests is missing or null, return an empty list with a 200 status
		if !exists || friendRequestsRaw == nil {
			return emails, 200, nil
		}

		friendRequests, ok := friendRequestsRaw.([]interface{})
		if !ok {
			return nil, 400, fmt.Errorf("failed to parse friendRequests") // HTTP 400 Bad Request
		}

		for _, request := range friendRequests {
			requestMap, ok := request.(map[string]interface{})
			if !ok {
				continue
			}
			toUserEmail, ok := requestMap["toUserEmail"].(string)
			if ok {
				emails = append(emails, toUserEmail)
			}
		}
		status = 200 // HTTP 200 OK
		return emails, status, nil

	default:
		return nil, 400, fmt.Errorf("invalid invitation type: %s", invitationType) // HTTP 400 Bad Request
	}
}
