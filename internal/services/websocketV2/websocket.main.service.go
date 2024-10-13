package websocketv2

import (
	"be_chat_app/models"
	"context"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
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

func (w *WebsocketServices) FetchAllMessageBoxes() {
	iter := w.FireStoreClient.Collection("messageBoxes").Documents(context.Background())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("failed to iterate: %v", err)
		}

		w.Hub.MessageBoxes[doc.Ref.ID] = &MessageBox{
			MessageBoxId: doc.Ref.ID,
			Clients:      make(map[string]*Client),
		}
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
