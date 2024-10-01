package websocket

import (
	"be_chat_app/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebsocketService struct {
	hub *Hub
}

func NewWebsocketService(h *Hub) *WebsocketService {
	return &WebsocketService{
		hub: h,
	}
}

type CreateRoomReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (ws *WebsocketService) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if _, ok := ws.hub.Rooms[req.ID]; ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room id already exists!!"})

		return
	}

	roomInfo := RoomInfo{
		ID:   req.ID,
		Name: req.Name,
	}

	ws.hub.Rooms[req.ID] = &Room{
		RoomInfo: roomInfo,
		Clients:  make(map[string]*Client),
	}

	ws.hub.RoomInfo <- &roomInfo

	c.JSON(http.StatusOK, req)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var id int = 1

func (ws *WebsocketService) JoinMasterRoom(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := &Client{
		Conn:     conn,
		Message:  make(chan *Message),
		ID:       strconv.Itoa(id),
		RoomID:   strconv.Itoa(id),
		Username: strconv.Itoa(id),
	}

	ws.hub.ClientOnMasterRoom <- client

	// message := &Message{
	// 	Content:  "User " + strconv.Itoa(id) + " has joined the master room",
	// 	RoomID:   strconv.Itoa(id),
	// 	Username: strconv.Itoa(id),
	// }

	// h.hub.Clients.Clients[strconv.Itoa(id)].Message <- message
	go client.readMessageOnMasterRoom(ws.hub)
	go client.writeMessageOnMasterRoom()
	fmt.Printf("Id %d", id)
	id++
}

func (ws *WebsocketService) JoinRoom(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roomID := c.Param("roomId")
	clientID := c.Query("userId")
	username := c.Query("username")

	cl := &Client{
		Conn:     conn,
		Message:  make(chan *Message, 10),
		ID:       clientID,
		RoomID:   roomID,
		Username: username,
	}

	m := &Message{
		Content:  "A new user has joined the room",
		RoomID:   roomID,
		Username: username,
	}

	ws.hub.Register <- cl
	ws.hub.Broadcast <- m

	go cl.writeMessage()
	cl.readMessage(ws.hub)
}

type RoomRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (ws *WebsocketService) GetRooms(c *gin.Context) {
	rooms := make([]RoomRes, 0)

	for _, r := range ws.hub.Rooms {
		rooms = append(rooms, RoomRes{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	c.JSON(http.StatusOK, rooms)
}

type ClientRes struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (ws *WebsocketService) GetClients(c *gin.Context) {
	var clients []ClientRes
	roomId := c.Param("roomId")

	if _, ok := ws.hub.Rooms[roomId]; !ok {
		clients = make([]ClientRes, 0)
		c.JSON(http.StatusOK, clients)
	}

	for _, c := range ws.hub.Rooms[roomId].Clients {
		clients = append(clients, ClientRes{
			ID:       c.ID,
			Username: c.Username,
		})
	}

	c.JSON(http.StatusOK, clients)
}

func (ws *WebsocketService) MakeFriend(c *gin.Context) {
	var req models.FriendRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	req.Status = "pending"
	req.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	ws.hub.MakingFriend <- &req

	c.JSON(http.StatusOK, gin.H{
		"request": req,
	})
}
