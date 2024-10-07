package websocketv2

import (
	"be_chat_app/models"

	"github.com/gorilla/websocket"
)

type ClientOnMasterRoom struct {
	Conn                     *websocket.Conn
	AcceptFriendNotification chan *models.Notification
	UserId                   string `json:"userId"`
	UserName                 string `json:"userName"`
}

func (c *ClientOnMasterRoom) WriteAcceptNotification() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.AcceptFriendNotification
		if !ok {
			return
		}

		c.Conn.WriteJSON(message)
	}
}

// func (c *ClientOnMasterRoom) readAcceptNotification(hub *Hub) {
// 	defer func() {
// 		c.Conn.Close()
// 	}()

// 	for {
// 		_, payload, err := c.Conn.ReadMessage()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				log.Printf("error: %v", err)
// 			}
// 			break
// 		}

// 		notification := &models.Notification{}
// 	}
// }