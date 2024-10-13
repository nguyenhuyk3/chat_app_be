package router

import (
	"be_chat_app/api"
	"be_chat_app/internal/services/user"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var UserServices *user.UserServices

func InitUserRouter(r *gin.Engine, userServices *user.UserServices) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	userApi := api.NewUserApi(userServices)

	r.GET("/users/search", userApi.SearchUserIdByEmail)
	r.GET("/users/get_user", userApi.GetUserByEmail)
	r.GET("/users/get_receiving_invitation_box", userApi.GetReceivingInvitationBox)
	r.GET("/users/get_sending_invitation_box", userApi.GetSendingInvitationBox)
	r.GET("/users/get_sub_ids", userApi.GetSubIds)
	r.GET("/users/get_all_message_boxes", userApi.GetAllMessageBoxesByUserId)
	r.GET("/users/get_message_box_by_id", userApi.GetMessageBoxById)
	r.POST("/users/make_friend", userApi.MakeFriend)
	// r.POST("/users/accept_friend", userApi.AcceptFriend)
	r.POST("/users/delete_friend_request_for_sending", userApi.DeleteFriendRequestForSending)
	r.POST("/users/delete_friend_request_for_receiving", userApi.DeleteFriendRequestForReceiving)
	r.POST("/users/update_information", userApi.UpdateInformation)
}
