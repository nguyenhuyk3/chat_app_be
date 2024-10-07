package api

import (
	"be_chat_app/internal/services/user"
	websocketv2 "be_chat_app/internal/services/websocketV2"
	"be_chat_app/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebsocketApi struct {
	WebsocketServices *websocketv2.WebsocketServices
	UserServices      *user.UserServices
}

func NewWebsocketApi(
	websocketServices *websocketv2.WebsocketServices,
	userServices *user.UserServices) *WebsocketApi {
	return &WebsocketApi{
		WebsocketServices: websocketServices,
		UserServices:      userServices,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (w *WebsocketApi) JoinMasterRoom(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.Query("user_id")

	client := &websocketv2.ClientOnMasterRoom{
		Conn:                     conn,
		AcceptFriendNotification: make(chan *models.Notification),
		UserId:                   userId,
	}

	w.WebsocketServices.Hub.ClientGetInToMasterRoom <- client

	go client.WriteAcceptNotification()
}

type AcceptFriendReq struct {
	FromUserEmail string `json:"fromUserEmail"`
	FromUserName  string `json:"fromUserName"`
	ToUserEmail   string `json:"toUserEmail"`
	ToUserName    string `json:"toUserName"`
}

func (w *WebsocketApi) AcceptFriend(c *gin.Context) {
	var req AcceptFriendReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request data"})
		return
	}

	fromUserId, toUserId, status, err := w.UserServices.AcceptFriend(req.ToUserEmail, req.FromUserEmail)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	firstUser := models.InforUser{
		Id:       fromUserId,
		Email:    req.FromUserEmail,
		FullName: req.FromUserName,
		Avtar:    "",
	}

	secondUser := models.InforUser{
		Id:       toUserId,
		Email:    req.ToUserEmail,
		FullName: req.ToUserName,
		Avtar:    "",
	}

	status, err = w.UserServices.AddMessageBoxForBothUser(firstUser, secondUser)
	if err != nil {
		c.JSON(status, gin.H{"error": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(status, gin.H{"message": "perform sucessfully"})

	notiSendFromUser := models.Notification{
		FromUserInfor: models.FromUserInfor{
			FromUserEmail: req.FromUserEmail,
			FromUserName:  req.FromUserName,
		},
		ToUserId:  toUserId,
		Content:   fmt.Sprintf("Từ giờ bạn và %s có thể trò chuyện", req.FromUserName),
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	notiSendToUser := models.Notification{
		FromUserInfor: models.FromUserInfor{
			FromUserEmail: req.ToUserEmail,
			FromUserName:  req.ToUserName,
		},
		ToUserId:  fromUserId,
		Content:   fmt.Sprintf("Từ giờ bạn và %s có thể trò chuyện", req.ToUserName),
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	w.WebsocketServices.Hub.AcceptFriendNotification <- &notiSendFromUser
	w.WebsocketServices.Hub.AcceptFriendNotification <- &notiSendToUser
}
