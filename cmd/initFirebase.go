package cmd

import (
	firebaseServices "be_chat_app/internal/firebase"
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
)

var firestoreClient *firestore.Client

func InitFirebase() (*firebase.App, *firestore.Client, *messaging.Client, error) {
	firebaseConfig := &firebaseServices.FirebaseConfig{
		CredentialsFile: "./chat-app-flutter-e4dda-firebase-adminsdk-qkgt3-4086323c29.json",
		ProjectId:       "chat-app-flutter-e4dda",
	}

	firebaseApp, err := firebaseServices.InitializeFirebaseApp(firebaseConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error initializing Firestore client: %w", err)
	}

	firebaseClient, err := firebaseServices.InitializeFirestoreClient(firebaseApp)
	if err != nil {
		log.Fatalf("error getting Firestore client: %v", err)
	}

	messagingClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		// Close Firestore client if Messaging client initialization fails
		firestoreClient.Close()
		return nil, nil, nil, err
	}

	firestoreClient = firebaseClient

	return firebaseApp, firestoreClient, messagingClient, nil
}

func CloseFirestoreClient() {
	if firestoreClient != nil {
		err := firestoreClient.Close()
		if err != nil {
			log.Printf("Error closing Firestore client: %v\n", err)
		}
	}
}
