package cmd

import (
	"be_chat_app/internal/firebase"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	firebaseApp "firebase.google.com/go"
)

var firestoreClient *firestore.Client

// InitFirebaseClient initializes Firebase app and Firestore client
func InitFirebase() (*firebaseApp.App, *firestore.Client, error) {
	firebaseConfig := &firebase.FirebaseConfig{
		CredentialsFile: "./chat-app-flutter-e4dda-firebase-adminsdk-qkgt3-4086323c29.json",
		ProjectId:       "chat-app-flutter-e4dda",
	}

	app, err := firebase.InitializeFirebaseApp(firebaseConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("error initializing Firestore client: %w", err)
	}

	client, err := firebase.InitializeFirestoreClient(app)
	if err != nil {
		log.Fatalf("error getting Firestore client: %v", err)
	}

	firestoreClient = client

	return app, client, nil
}

func CloseFirestoreClient() {
	if firestoreClient != nil {
		err := firestoreClient.Close()
		if err != nil {
			log.Printf("Error closing Firestore client: %v\n", err)
		}
	}
}
