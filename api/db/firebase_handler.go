package db

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/HeavenAQ/api/secret"
	"google.golang.org/api/option"
)

func NewFirebaseHandler() (*FirebaseHandler, error) {
	ctx := context.Background()

	// get firebase credentials from secret manager
	secretName := secret.GetSecretNameString(os.Getenv("FIREBASE_CREDENTIALS"))
	firebaseCredentials, err := secret.AccessSecretVersion(secretName)

	// get service account and initialize firebase app
	sa := option.WithCredentialsJSON(firebaseCredentials)
	conf := &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}
	app, err := firebase.NewApp(ctx, conf, sa)
	if err != nil {
		return nil, err
	}

	// initialize firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatal("Error initializing firebase database client:", err)
	}
	return &FirebaseHandler{client, ctx}, nil
}

func (handler *FirebaseHandler) GetUsersCollection() *firestore.CollectionRef {
	collection := os.Getenv("FIREBASE_USERS")
	return handler.dbClient.Collection(collection)
}
