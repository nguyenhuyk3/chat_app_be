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
	userId := c.Query("userId")

	messageBoxes, status, err := u.UserServices.GetAllMessageBoxesByUserId(userId)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messageBoxesReponse": messageBoxes})
}
