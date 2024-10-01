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

	// Channel to store whether the user exists or not
	existsChan := make(chan bool, 1)

	// We have 2 work to complete
	wg.Add(2)

	// Check if user exists
	go func() {
		// Decreasing counter by 1 when this goroutine finishes
		defer wg.Done()

		exists, status, err := u.UserServices.CheckUserIfExistByEmail(req.ToUserEmail)
		if err != nil {
			mu.Lock()
			finalErr = err
			finalStatus = status
			mu.Unlock()

			return
		}

		if !exists {
			mu.Lock()
			finalErr = fmt.Errorf("to user not found")
			finalStatus = http.StatusNotFound
			mu.Unlock()

			existsChan <- false

			return
		}

		existsChan <- true
	}()

	// Get infor of sender
	go func() {
		defer wg.Done()

		fUser, status, err := u.UserServices.GetUserByEmail(req.FromUserInfor.FromUserEmail)
		if err != nil {
			mu.Lock()
			finalErr = err
			finalStatus = status
			mu.Unlock()

			return
		}

		mu.Lock()
		fromUser = fUser
		mu.Unlock()
	}()

	wg.Wait()

	if finalErr != nil {
		c.JSON(finalStatus, gin.H{"error": fmt.Sprintf("%v", finalErr)})

		return
	}

	if !<-existsChan {
		c.JSON(http.StatusNotFound, gin.H{"error": "to user not found"})

		return
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		tUser, status, err := u.UserServices.GetUserByEmail(req.ToUserEmail)
		if err != nil {
			finalStatus = status
			finalErr = err

			return
		}

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

		status, err := u.UserServices.AddFriendReqToBox("sendingInvitationBoxes", fromUser.Id, fromUser.SendingInvitationBox, req)
		errChan <- err
		statusChan <- status
	}()

	go func() {
		defer wg.Done()

		status, err := u.UserServices.AddFriendReqToBox("receivingInvitationBoxes", toUser.Id, toUser.ReceivingInvitationBox, req)
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

	user, status, err := u.UserServices.GetUserByEmail(email)
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
	ToUserEmail   string `json:"toUserEmail"`
	FromUserEmail string `json:"fromUserEmail"`
}

func (u *UserApi) AcceptFriend(c *gin.Context) {
	var req AddFriendReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})

		return
	}

	status, err := u.UserServices.AcceptFriend(req.ToUserEmail, req.FromUserEmail)
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

	c.JSON(http.StatusOK, gin.H{"make_friend_request": makeFriendReq})
}

func (u *UserApi) GetSendingInvitationBox(c *gin.Context) {
	invitationBoxId := c.Query("sending_invitation_box")

	makeFriendReq, status, err := u.UserServices.GetSendingInvitationBox(invitationBoxId)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})

		return
	}

	c.JSON(http.StatusOK, gin.H{"make_friend_request": makeFriendReq})
}
