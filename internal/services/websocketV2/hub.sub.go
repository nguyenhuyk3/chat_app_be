package websocketv2

import (
	"be_chat_app/internal/services/notification"
	"be_chat_app/models"
	"fmt"
	"log"
	"time"

	"github.com/pion/webrtc/v3"
)

func (h *Hub) broadcastUserStatus(userId string, isOnline bool) []string {
	friends := h.UserServices.GetFriendIdsById(userId)
	newTokenDevice := ""

	if isOnline {
		newTokenDevice, _, _ = h.NotificationServices.GetTokenByUserId(userId)
	}

	newUserStatus := &UserStatus{
		UserId:      userId,
		IsOnline:    isOnline,
		TokenDevice: newTokenDevice,
	}

	for _, friendId := range friends {
		if friendRoom, ok := h.MasterRooms[friendId]; ok {
			friendRoom.ClientOnMasterRoom.UserStatus <- newUserStatus
		}
	}
	return friends
}

func (h *Hub) getOnlineFriends(friendIds []string) []string {
	onlineFriends := []string{}

	for _, friendId := range friendIds {
		if _, ok := h.MasterRooms[friendId]; ok {
			onlineFriends = append(onlineFriends, friendId)
		}
	}
	return onlineFriends
}

func (h *Hub) handleClientGetIntoMasterRoom(client *ClientOnMasterRoom) {
	userId := client.UserId
	if _, ok := h.MasterRooms[userId]; !ok {
		newMasterRoom := NewMasterRoom(client)

		h.MasterRooms[userId] = newMasterRoom

		friendIds := h.broadcastUserStatus(userId, true)
		onlineFriends := h.getOnlineFriends(friendIds)

		client.SendOnlineFriends(onlineFriends)
	}
}

func (h *Hub) handleClientGetOutMasterRoom(client *ClientOnMasterRoom) {
	userId := client.UserId

	go func() {
		_, _ = h.NotificationServices.DeleteToken(userId)
	}()

	if _, ok := h.MasterRooms[userId]; ok {
		defer func() {
			delete(h.MasterRooms, userId)
			close(client.AcceptFriendNotification)
			close(client.UserStatus)
			close(client.LastStateForMessageBoxOnMasterRoom)
		}()
		h.broadcastUserStatus(userId, false)
	}
}

func (h *Hub) handleAcceptFriendNotification(notification *models.Notification) {
	userId := notification.ToUserId
	if room, ok := h.MasterRooms[userId]; ok {
		room.ClientOnMasterRoom.AcceptFriendNotification <- notification
	}
}

func (h *Hub) handleClientGetInMessageBox(client *Client) {
	messageBoxId := client.MessageBoxId
	userId := client.UserId

	if _, ok := h.MessageBoxes[messageBoxId].Clients[userId]; !ok {
		h.MessageBoxes[messageBoxId].Clients[userId] = client

		// If two clients are in the message box, send a notification
		if len(h.MessageBoxes[messageBoxId].Clients) == 2 {
			for key := range h.MessageBoxes[messageBoxId].Clients {
				if key != userId {
					newMessage := &models.CommingMessage{
						SenderId:     userId,
						MessageBoxId: messageBoxId,
						Content:      "anhiuemlove33333!@#@#@!!!****&(*&@(^&*()concak",
						State:        "",
					}
					h.ReadedMessageNotification <- newMessage
					break
				}
			}
		}
	}
}

func (h *Hub) handleClientGetOutMessageBox(client *Client) {
	messageBoxId := client.MessageBoxId
	userId := client.UserId

	if _, ok := h.MessageBoxes[messageBoxId].Clients[userId]; ok {
		delete(h.MessageBoxes[messageBoxId].Clients, userId)
		close(client.Message)
	}
}

func (h *Hub) handleBroadcastMessage(message *models.CommingMessage) {
	messageBoxId := message.MessageBoxId

	if box, ok := h.MessageBoxes[messageBoxId]; ok {
		newState := "chưa đọc"
		if len(box.Clients) == 2 {
			newState = "đã đọc"
		}

		newMessage := &models.Message{
			SenderId:  message.SenderId,
			Type:      message.Type,
			Content:   message.Content,
			SendedId:  message.SendedId,
			State:     newState,
			CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		lastMessageState := message.Content

		if message.Type != "text" {
			lastMessageState = "Video"
		}

		lastStateForMessageBoxOnMasterRoomForSender := &LastStateForMessageBoxOnMasterRoom{
			IsLastStateStruct: true,
			SenderId:          message.SenderId,
			MessageBoxId:      messageBoxId,
			LastMessage:       lastMessageState,
			LastTime:          message.CreatedAt.Format("2006-01-02 15:04:05"),
			LastStatus:        "đã đọc",
		}
		lastStateForMessageBoxOnMasterRoomForReiceiver := &LastStateForMessageBoxOnMasterRoom{
			IsLastStateStruct: true,
			SenderId:          message.SenderId,
			MessageBoxId:      messageBoxId,
			LastMessage:       lastMessageState,
			LastTime:          message.CreatedAt.Format("2006-01-02 15:04:05"),
			LastStatus:        newState,
		}
		receiverName, _, _ := h.UserServices.GetFullNameById(message.SenderId)

		for _, client := range box.Clients {
			if len(box.Clients) == 1 {
				newNotification := notification.MessageNotification{
					Token:  message.TokenDevice,
					Avatar: "",
					Title:  box.Clients[message.SenderId].FullName,
					Body:   "1 tin nhắn mới",
				}
				go func() {
					notification.SendNotificationForCommingMessage(h.MessagingClient, newNotification,
						messageBoxId, message.ReceiverId, message.TokenDevice, receiverName)
				}()
			}
			var isFirst bool = true
			if isFirst {
				isFirst = false
				fmt.Println(client.UserId)
				fmt.Println(message.ReceiverId)
				if clientOnMasterRoom, ok := h.MasterRooms[message.ReceiverId]; ok {
					clientOnMasterRoom.ClientOnMasterRoom.LastStateForMessageBoxOnMasterRoom <- lastStateForMessageBoxOnMasterRoomForReiceiver
				}
				if clientOnMasterRoom, ok := h.MasterRooms[client.UserId]; ok {
					clientOnMasterRoom.ClientOnMasterRoom.LastStateForMessageBoxOnMasterRoom <- lastStateForMessageBoxOnMasterRoomForSender
				}
			}
			client.Message <- newMessage
		}
	}
}

func (h *Hub) handleReadedMessageNotification(readedMessage *models.CommingMessage) {
	userId := readedMessage.SenderId
	messageBoxId := readedMessage.MessageBoxId
	content := readedMessage.Content

	if box, ok := h.MessageBoxes[messageBoxId]; ok {
		newMessage := &models.Message{
			SenderId:  userId,
			Content:   content,
			State:     "",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		for _, client := range box.Clients {
			client.Message <- newMessage
		}
	}
}

func (h *Hub) HandleCallRequest(caller *ClientOnMasterRoom, receiverId string) {
	receiver := h.MasterRooms[receiverId]
	if receiver.ClientOnMasterRoom == nil {
		log.Println("Receiver not online")
		return
	}
	// Create PeerConnection for caller
	var err error
	caller.PeerConnection, err = h.SetupPeerConnection()
	if err != nil {
		log.Println("Error setting up peer connection:", err)
		return
	}
	// Set OnICECandidate for the caller
	caller.PeerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			caller.SendICECandidate(c.ToJSON())
		}
	})
	// Create SDP offer
	offer, err := caller.PeerConnection.CreateOffer(nil)
	if err != nil {
		log.Println("Error creating offer:", err)
		return
	}
	// Set Remote description for caller
	err = caller.PeerConnection.SetLocalDescription(offer)
	if err != nil {
		log.Println("Error setting local description:", err)
		return
	}
	// Send SDP offer to recipient
	message := VideoMessage{
		Type:        "call_offer",
		SessionDesc: &offer,
	}
	receiver.ClientOnMasterRoom.sendMessage(message)
	// Handle call status
	go handleCallStatus(caller, receiver.ClientOnMasterRoom)
}

func handleCallStatus(caller *ClientOnMasterRoom, receiver *ClientOnMasterRoom) {
	time.Sleep(20 * time.Second)

	// Kiểm tra xem receiver đã nhận cuộc gọi chưa
	if receiver.PeerConnection == nil {
		log.Println("Call missed for recipient:", receiver.UserId)
	}
}
