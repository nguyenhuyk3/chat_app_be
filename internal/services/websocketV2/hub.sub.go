package websocketv2

import (
	"be_chat_app/internal/services/notification"
	"be_chat_app/models"
	"time"
)

func (h *Hub) handleClientGetIntoMasterRoom(client *ClientOnMasterRoom) {
	userId := client.UserId
	if _, ok := h.MasterRooms[userId]; !ok {
		newMasterRoom := NewMasterRoom(client)
		h.MasterRooms[userId] = newMasterRoom
	}
}

func (h *Hub) handleClientGetOutMasterRoom(client *ClientOnMasterRoom) {
	userId := client.UserId
	if _, ok := h.MasterRooms[userId]; ok {
		delete(h.MasterRooms, userId)
		close(client.AcceptFriendNotification)
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

		for _, client := range box.Clients {
			if len(box.Clients) == 1 {
				newNotification := notification.MessageNotification{
					Token:  message.TokenDevice,
					Avatar: "",
					Title:  box.Clients[message.SenderId].FullName,
					Body:   "1 tin nhắn mới",
				}

				go func() {
					notification.SendNotificationForCommingMessage(h.MessagingClient, newNotification)
				}()
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
