package firebase

import (
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// FirebaseConfig stores the necessary configuration to initialize Firebase
type FirebaseConfig struct {
	CredentialsFile string
	ProjectId       string
}

// InitializeFirebaseApp initializes a Firebase app with the given credentials
func InitializeFirebaseApp(config *FirebaseConfig) (*firebase.App, error) {
	opt := option.WithCredentialsFile(config.CredentialsFile)
	conf := &firebase.Config{
		ProjectID: config.ProjectId,
	}
	app, err := firebase.NewApp(context.Background(), conf, opt)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// InitializeFirestoreClient initializes a Firestore client using the Firebase app
func InitializeFirestoreClient(app *firebase.App) (*firestore.Client, error) {
	client, err := app.Firestore(context.Background())
	if err != nil {
		return nil, err
	}

	return client, nil
}
