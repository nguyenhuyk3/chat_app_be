package api

import (
	"be_chat_app/internal/services/user"
	"be_chat_app/models"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type UserApi struct {
	UserServices *user.UserServices
}

func NewUserApi(userServices *user.UserServices) *UserApi {
	return &UserApi{
		UserServices: userServices,
	}
}

func (u *UserApi) SearchUserIdByEmail(c *gin.Context) {
	var email string = c.Query("email")

	user, status, err := u.UserServices.SearchUserIdByEmail(email)
	if err != nil {
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"response": user,
	})
}

func (u *UserApi) MakeFriend(c *gin.Context) {
	var req models.FriendRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var finalErr error
	var finalStatus int
	var fromUser, toUser models.User
	var fromUserId, toUserId string

	wg.Add(1)
	go func() {
		defer wg.Done()

		fUser, fFromUserId, status, err := u.UserServices.GetUserByEmail(req.FromUserInfor.FromUserEmail)
		if err != nil {
			mu.Lock()
			finalErr = err
			finalStatus = status
			mu.Unlock()
			return
		}
		fromUserId = fFromUserId
		fromUser = fUser
	}()
	wg.Wait()

	if finalErr != nil {
		c.JSON(finalStatus, gin.H{"error": fmt.Sprintf("%v", finalErr)})
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		tUser, tUserId, status, err := u.UserServices.GetUserByEmail(req.ToUserEmail)
		if err != nil {
			finalStatus = status
			finalErr = err
			return
		}
		toUserId = tUserId
		toUser = tUser
	}()
	wg.Wait()

	if finalErr != nil {
		c.JSON(finalStatus, gin.H{"error": fmt.Sprintf("%v", finalErr)})
		return
	}

	errChan := make(chan error, 2)
	statusChan := make(chan int, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()

		status, err := u.UserServices.AddFriendReqToBox("sendingInvitationBoxes", fromUserId, fromUser.SendingInvitationBoxId, req)
		errChan <- err
		statusChan <- status
	}()
	go func() {
		defer wg.Done()

		status, err := u.UserServices.AddFriendReqToBox("receivingInvitationBoxes", toUserId, toUser.ReceivingInvitationBoxId, req)
		errChan <- err
		statusChan <- status
	}()
	wg.Wait()

	close(errChan)
	close(statusChan)

	for err := range errChan {
		if err != nil {
			finalErr = err
			finalStatus = <-statusChan
			break
		}
	}

	if finalErr != nil {
		c.JSON(finalStatus, gin.H{"error": fmt.Sprintf("%v", finalErr)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "perform successfully"})
}

func (u *UserApi) GetUserByEmail(c *gin.Context) {
	var email string = c.Query("email")

	user, _, status, err := u.UserServices.GetUserByEmail(email)
	if err != nil {
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(status, gin.H{
		"user": user,
	})
}

type AddFriendReq struct {
	FromUserEmail string `json:"fromUserEmail"`
	ToUserEmail   string `json:"toUserEmail"`
}

func (u *UserApi) AcceptFriend(c *gin.Context) {
	var req AddFriendReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	_, _, status, err := u.UserServices.AcceptFriend(req.ToUserEmail, req.FromUserEmail)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "perform sucessfully"})
}

func (u *UserApi) GetReceivingInvitationBox(c *gin.Context) {
	invitationBoxId := c.Query("receiving_invitation_box")

	makeFriendReq, status, err := u.UserServices.GetReceivingInvitationBox(invitationBoxId)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"makeFriendRequests": makeFriendReq})
}

func (u *UserApi) GetSendingInvitationBox(c *gin.Context) {
	invitationBoxId := c.Query("sending_invitation_box")

	makeFriendReq, status, err := u.UserServices.GetSendingInvitationBox(invitationBoxId)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"makeFriendRequests": makeFriendReq})
}

func (u *UserApi) GetSubIds(c *gin.Context) {
	email := c.Query("email")

	subIds, status, err := u.UserServices.GetSubIdsByEmail(email)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":                 email,
		"userId":                subIds.UserId,
		"receivingInvitationId": subIds.ReceivingInvitationBoxId,
		"sendingInvitationId":   subIds.SendingInvitationBoxId,
	})
}

type DeleteFriendReq struct {
	InvitationBoxId string `json:"invitationBoxId"`
	FromUserEmail   string `json:"fromUserEmail"`
	ToUserEmail     string `json:"toUserEmail"`
}

func (u *UserApi) DeleteFriendRequestForSending(c *gin.Context) {
	var req DeleteFriendReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	var fromUserBoxType = "sendingInvitationBoxes"
	var toUserBoxType = "receivingInvitationBoxes"

	status, err := u.UserServices.DeleteFriendRequest(fromUserBoxType, toUserBoxType,
		req.InvitationBoxId,
		req.FromUserEmail, req.ToUserEmail)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "perform sucessfully",
	})
}

func (u *UserApi) DeleteFriendRequestForReceiving(c *gin.Context) {
	var req DeleteFriendReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})

		return
	}

	var fromUserBoxType = "receivingInvitationBoxes"
	var toUserBoxType = "sendingInvitationBoxes"

	status, err := u.UserServices.DeleteFriendRequest(fromUserBoxType, toUserBoxType,
		req.InvitationBoxId,
		req.FromUserEmail, req.ToUserEmail)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "perform sucessfully",
	})
}

func (u *UserApi) GetAllMessageBoxesByUserId(c *gin.Context) {
	userId := c.Query("user_id")

	messageBoxes, status, err := u.UserServices.GetAllMessageBoxesByUserId(userId)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"messageBoxesReponse": messageBoxes})
}

func (u *UserApi) GetMessageBoxById(c *gin.Context) {
	messageBoxId := c.Query("message_box_id")

	messageBox, status, err := u.UserServices.GetMessageBoxById(messageBoxId)
	if err != nil {
		c.JSON(status, gin.H{"error": err})
	}
	c.JSON(status, gin.H{"messageBox": messageBox})
}

type UpdateInformationReq struct {
	UserId     string `json:"userId"`
	FullName   string `json:"fullName"`
	Genre      string `json:"genre"`
	DayOfBirth string `json:"dayOfBirth"`
}

func (u *UserApi) UpdateInformation(c *gin.Context) {
	var req UpdateInformationReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	newInformation := models.Information{
		FullName:   req.FullName,
		Genre:      req.Genre,
		DayOfBirth: req.DayOfBirth,
	}

	resultCh := make(chan struct {
		status int
		err    error
	}, 2)

	go func() {
		status, err := u.UserServices.ChangeInformationAtRoot(req.UserId, newInformation)
		resultCh <- struct {
			status int
			err    error
		}{status: status, err: err}
	}()
	go func() {
		status, err := u.UserServices.ChangeInformationAtMessageBoxes(req.UserId, newInformation)
		resultCh <- struct {
			status int
			err    error
		}{status: status, err: err}
	}()

	var errors []string
	var finalStatus int = http.StatusOK

	for i := 0; i < 2; i++ {
		result := <-resultCh
		if result.err != nil {
			errors = append(errors, result.err.Error())
			finalStatus = result.status
		}
	}
	if len(errors) > 0 {
		c.JSON(finalStatus, gin.H{"errors": errors})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "perform successfully"})
}

type ReadUnreadedMessagesReq struct {
	UserId       string `json:"userId"`
	MessageBoxId string `json:"messageBoxId"`
}

func (u *UserApi) ReadUnreadedMessages(c *gin.Context) {
	var req ReadUnreadedMessagesReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	status, err := u.UserServices.ReadUnreadedMessages(req.MessageBoxId, req.UserId)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(status, gin.H{"message": "perform successfully"})
}

type UpdateMessageBySendedIdReq struct {
	MessageBoxId string `json:"messageBoxId"`
	SendedId     string `json:"sendedId"`
	Content      string `json:"content"`
}

func (u *UserApi) UpdateMessageBySendedId(c *gin.Context) {
	var req UpdateMessageBySendedIdReq
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	status, err := u.UserServices.UpdateMessageBySendedId(req.MessageBoxId, req.SendedId, req.Content)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(status, gin.H{"message": "perform successfully"})
}

func (u *UserApi) GetAllFriendEmailsById(c *gin.Context) {
	userId := c.Query("user_id")
	friendEmails, status, _ := u.UserServices.GetFriendEmailsById(userId)

	c.JSON(status, gin.H{"friendEmails": friendEmails})
}

func (u *UserApi) GetInformationByEmail(c *gin.Context) {
	email := c.Query("email")
	information, status, err := u.UserServices.GetInformationByEmail(email)

	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"information": information})
}

func (u *UserApi) GetEmailsFromInvitationBox(c *gin.Context) {
	invitationBoxId := c.Query("invitation_box_id")
	invitationBoxType := c.Query("invitation_box_type")

	emails, status, err := u.UserServices.GetEmailsFromInvitationBox(invitationBoxId, invitationBoxType)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"emails": emails})
}

func (u *UserApi) GetFullNameById(c *gin.Context) {
	userId := c.Query("user_id")
	fullName, status, err := u.UserServices.GetFullNameById(userId)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"fullName": fullName})
}
