package api

import (
	"be_chat_app/internal/services/notification"
	"be_chat_app/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationApi struct {
	NotificationServices *notification.NotificationServices
}

func NewNotificationApi(notificationServices *notification.NotificationServices) *NotificationApi {
	return &NotificationApi{
		NotificationServices: notificationServices,
	}
}

func (n *NotificationApi) SaveToken(c *gin.Context) {
	var req models.UserToken
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	status, err := n.NotificationServices.SaveToken(req)
	if err != nil {
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "perform sucessfully"})
}

func (n *NotificationApi) GetTokenByUserId(c *gin.Context) {
	userId := c.Query("user_id")
	token, status, err := n.NotificationServices.GetTokenByUserId(userId)
	if err != nil {
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
