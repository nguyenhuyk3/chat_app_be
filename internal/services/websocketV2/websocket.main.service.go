package websocketv2

import (
	"be_chat_app/models"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
)

type WebsocketServices struct {
	Hub             *Hub
	FireStoreClient *firestore.Client
}

func NewWebsocketService(h *Hub, fireStoreClient *firestore.Client) *WebsocketServices {
	return &WebsocketServices{
		Hub:             h,
		FireStoreClient: fireStoreClient,
	}
}

func (w *WebsocketServices) ProcessCommingMessages() {
	var commingMessageBatch []models.CommingMessage
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case commingMessage := <-w.Hub.CommingMessage:
			commingMessageBatch = append(commingMessageBatch, *commingMessage)

			sort.Slice(commingMessageBatch, func(i, j int) bool {
				return commingMessageBatch[i].CreatedAt.Before(commingMessageBatch[j].CreatedAt)
			})

			if len(commingMessageBatch) >= 10 {
				_, err := w.SaveBatchMessages(commingMessageBatch)
				if err != nil {
					log.Fatalf("error when saving batch: %v", err)
				}

				commingMessageBatch = []models.CommingMessage{}
			}
		case <-ticker.C:
			if len(commingMessageBatch) > 0 {
				_, err := w.SaveBatchMessages(commingMessageBatch)
				if err != nil {
					log.Fatalf("error when saving batch: %v", err)
				}

				commingMessageBatch = []models.CommingMessage{}
			}
		}
	}
}
