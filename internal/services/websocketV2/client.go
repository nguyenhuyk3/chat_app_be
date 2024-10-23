package websocketv2

import (
	"be_chat_app/internal/services/user"
	"be_chat_app/models"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn         *websocket.Conn
	UserServices *user.UserServices
	Message      chan *models.Message
	MessageBoxId string `json:"messageBoxId"`
	UserId       string `json:"userId"`
	FullName     string `json:"fullName"`
}

type MessageBox struct {
	MessageBoxId string `json:"messageBoxId"`
	Clients      map[string]*Client
}

func (c *Client) WriteMessage() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		message, ok := <-c.Message
		if !ok {
			return
		}
		c.Conn.WriteJSON(message)
	}
}

// func saveFile(filename string, chunks map[int][]byte) error {
// 	filePath := filepath.Join("./assets/videos", filename)
// 	file, err := os.Create(filePath)
// 	if err != nil {
// 		return fmt.Errorf("error creating file: %v", err)
// 	}
// 	defer file.Close()

// 	// Ghi từng chunk vào file theo thứ tự
// 	for i := 1; i <= len(chunks); i++ {
// 		if _, err := file.Write(chunks[i]); err != nil {
// 			return fmt.Errorf("error writing to file: %v", err)
// 		}
// 	}
// 	return nil
// }

func (c *Client) ReadMessage(hub *Hub) {
	defer func() {
		hub.ClientGetOutMessageBox <- c
		c.Conn.Close()
	}()

	var incomingData struct {
		SenderId     string `json:"senderId"`
		Token        string `json:"token"`
		ReceiverId   string `json:"receiverId"`
		MessageBoxId string `json:"messageBoxId"`
		SendedId     string `json:"sendedId,omitempty"`
		Type         string `json:"type"`
		Content      string `json:"content"`
	}

	for {
		_, content, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error (ReadMessage): %v", err)
			}
			break
		}

		if err := json.Unmarshal(content, &incomingData); err != nil {
			log.Printf("error unmarshalling message: %v", err)
			continue
		}

		sendedId := incomingData.SendedId
		if incomingData.Type == "text" {
			sendedId = ""
		}

		commingMessage := &models.CommingMessage{
			MessageBoxId: incomingData.MessageBoxId,
			SenderId:     incomingData.SenderId,
			TokenDevice:  incomingData.Token,
			ReceiverId:   incomingData.ReceiverId,
			Type:         incomingData.Type,
			Content:      incomingData.Content,
			SendedId:     sendedId,
			State:        "chưa đọc",
			CreatedAt:    time.Now(),
		}
		hub.Broadcast <- commingMessage
		hub.CommingMessage <- commingMessage
	}
}

// type FileInfo struct {
// 	Name string `json:"name"`
// 	Size int64  `json:"size"`
// }

// func (c *Client) ReadFile(h *Hub) {
// 	defer c.Conn.Close()
// 	var file *os.File
// 	var currentFileName string

// 	for {
// 		messageType, r, err := c.Conn.NextReader()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				log.Printf("Error: %v", err)
// 			}
// 			break
// 		}

// 		if messageType == websocket.BinaryMessage {
// 			var metadata struct {
// 				SenderId     string `json:"senderId"`
// 				ReceiverId   string `json:"receiverId"`
// 				MessageBoxId string `json:"messageBoxId"`
// 				Type         string `json:"type"`
// 				FileName     string `json:"fileName"`
// 				ChunkIndex   int    `json:"chunkIndex"`
// 				IsLastChunk  bool   `json:"isLastChunk"`
// 			}

// 			metadataBytes, err := io.ReadAll(io.LimitReader(r, 2048)) // Tăng giới hạn đọc lên 2KB
// 			if err != nil {
// 				log.Printf("Error reading metadata: %v\n", err)
// 				continue
// 			}

// 			err = json.Unmarshal(metadataBytes, &metadata)
// 			if err != nil {
// 				log.Printf("Error decoding file info: %v\n", err)
// 				continue
// 			}

// 			if metadata.FileName != currentFileName {
// 				if file != nil {
// 					file.Close()
// 				}
// 				currentFileName = metadata.FileName
// 				filePath := filepath.Join("assets/videos", currentFileName)
// 				file, err = os.Create(filePath)
// 				if err != nil {
// 					log.Printf("Error creating file: %v\n", err)
// 					continue
// 				}
// 			}

// 			_, err = io.Copy(file, r)
// 			if err != nil {
// 				log.Printf("Error copying chunk content: %v\n", err)
// 				continue
// 			}

// 			// Gửi xác nhận cho client
// 			if err := c.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Chunk %d received", metadata.ChunkIndex))); err != nil {
// 				log.Println(err)
// 				break
// 			}

// 			if metadata.IsLastChunk {
// 				file.Close()
// 				file = nil
// 				currentFileName = ""
// 				log.Printf("File %s uploaded successfully\n", metadata.FileName)
// 			}
// 		}
// 	}

// 	if file != nil {
// 		file.Close()
// 	}
// }
