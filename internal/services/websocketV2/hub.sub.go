package websocketv2

import (
	"be_chat_app/internal/services/notification"
	"be_chat_app/models"
	"fmt"
	"time"
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
						Type:         "join-message-box",
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

	if message.Type == "video" || message.Type == "text" || message.Type == "audio" {
		fmt.Println(message.Type)
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

			if message.Type == "video" {
				lastMessageState = "Video anhiuemlove33333!@#@#@!!!****&(*&@(^&*()concak"
			} else if message.Type == "audio" {
				lastMessageState = "Audio anhiuemlove33333!@#@#@!!!****&(*&@(^&*()concak"
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

			for _, client := range box.Clients {
				if len(box.Clients) == 1 {
					receiverName, _, _ := h.UserServices.GetFullNameById(message.SenderId)
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
	} else if message.Type == "completed-media-call" || message.Type == "missed-media-call" || message.Type == "declined-media-call" {
		if box, ok := h.MessageBoxes[messageBoxId]; ok {
			state := "đã đọc"

			if message.Type == "missed-media-call" && len(box.Clients) == 1 {
				state = "chưa đọc"
			}
			newMessage := &models.Message{
				SenderId:  message.SenderId,
				Type:      message.Type,
				Content:   message.Content,
				State:     state,
				CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
			}

			for _, client := range box.Clients {
				client.Message <- newMessage
			}
		}
	} else {
		if message.Type == "declined-media-call-at-foreground" {
			commingMessage := &models.CommingMessage{
				MessageBoxId: message.MessageBoxId,
				SenderId:     message.SenderId,
				TokenDevice:  message.TokenDevice,
				ReceiverId:   message.ReceiverId,
				CallType:     message.CallType,
				Type:         "declined-media-call",
				Content:      message.Content,
				Sdp:          message.Sdp,
				Candidate:    message.Candidate,
				SendedId:     message.SendedId,
				State:        "chưa đọc",
				CreatedAt:    time.Now(),
			}
			h.CommingMessage <- commingMessage

			if box, ok := h.MessageBoxes[messageBoxId]; ok {
				state := "đã đọc"
				if len(box.Clients) == 1 {
					state = "chưa đọc"
				}
				newMessageSignal := &models.Message{
					SenderId:  message.SenderId,
					Type:      "declined-media-call-signal",
					Content:   message.Content,
					State:     state,
					CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
				}
				newMessage := &models.Message{
					SenderId:  message.SenderId,
					Type:      "declined-media-call",
					Content:   message.Content,
					State:     state,
					CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
				}
				for _, client := range box.Clients {
					client.Message <- newMessageSignal
					client.Message <- newMessage
					if len(box.Clients) == 1 {
						lastMessageState := "Cuộc gọi đã bị từ chối"
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
							LastStatus:        "chưa đọc",
						}
						var isFirst bool = true
						if isFirst {
							isFirst = false
							if clientOnMasterRoom, ok := h.MasterRooms[message.ReceiverId]; ok {
								clientOnMasterRoom.ClientOnMasterRoom.LastStateForMessageBoxOnMasterRoom <- lastStateForMessageBoxOnMasterRoomForReiceiver
							}
							if clientOnMasterRoom, ok := h.MasterRooms[client.UserId]; ok {
								clientOnMasterRoom.ClientOnMasterRoom.LastStateForMessageBoxOnMasterRoom <- lastStateForMessageBoxOnMasterRoomForSender
							}
						}
					}
				}
			}
		} else {
			if box, ok := h.MessageBoxes[messageBoxId]; ok {
				for _, client := range box.Clients {
					if client.UserId != message.SenderId {
						newMessage := &models.Message{
							SenderId:  message.SenderId,
							CallType:  message.CallType,
							Type:      message.Type,
							Content:   message.Content,
							Sdp:       message.Sdp,
							Candidate: message.Candidate}

						client.Message <- newMessage
						break
					} else if len(box.Clients) == 1 && (message.Type == "offer" ||
						message.Type == "answer" ||
						message.Type == "ice-candidate" ||
						message.Type == "declined-media-call-signal") {
						fmt.Println("========================================================")
						fmt.Println(message.Type + " " + message.SenderId)
						fmt.Println("========================================================")
						// receiverName, _, _ := h.UserServices.GetFullNameById(message.SenderId)
						// newNotification := notification.MessageNotification{
						// 	Token:  message.TokenDevice,
						// 	Avatar: "",
						// 	Title:  box.Clients[message.SenderId].FullName,
						// 	Type:   message.Type,
						// 	Body:   "1 tin nhắn mới",
						// }
						newOffer := &OfferNotification{
							MessageBoxId: messageBoxId,
							SenderId:     message.SenderId,
							SenderName:   box.Clients[message.SenderId].FullName,
							Sdp:          message.Sdp,
							Candidate:    message.Candidate,
							CallType:     message.CallType,
							Type:         message.Type,
							Token:        message.TokenDevice,
						}
						if receiver, ok := h.MasterRooms[message.ReceiverId]; ok {
							receiver.ClientOnMasterRoom.OfferNotification <- newOffer
						}
						// go func() {
						// 	notification.SendNotificationForCommingMessage(h.MessagingClient, newNotification,
						// 		messageBoxId, message.ReceiverId, message.TokenDevice, receiverName)
						// }()
					}
				}
			}
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
			Type:      readedMessage.Type,
			Content:   content,
			State:     "",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		for _, client := range box.Clients {
			client.Message <- newMessage
		}
	}
}
