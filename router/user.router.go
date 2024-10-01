package router

import (
	"be_chat_app/api"
	"be_chat_app/internal/services/user"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitUserRouter(r *gin.Engine, firebaseClient *firestore.Client) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	userServices := user.NewUserServices(firebaseClient)
	userApi := api.NewUserApi(userServices)

	r.GET("/users/search", userApi.SearchUserIdByEmail)
	r.GET("/users/get_user", userApi.GetUserByEmail)
	r.GET("/users/get_receiving_invitation_box", userApi.GetReceivingInvitationBox)
	r.GET("/users/get_sending_invitation_box", userApi.GetSendingInvitationBox)
	r.POST("/users/make_friend", userApi.MakeFriend)
	r.POST("/users/accept_friend", userApi.AcceptFriend)
}
