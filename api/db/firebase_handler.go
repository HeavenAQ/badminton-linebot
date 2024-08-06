package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func NewFirebaseHandler() (*FirebaseHandler, error) {
	ctx := context.Background()
	fmt.Printf("Firebase credentials file: %s\n", os.Getenv("FIREBASE_CREDENTIAL"))
	sa := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIAL"))
	conf := &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}
	app, err := firebase.NewApp(ctx, conf, sa)
	if err != nil {
		return nil, err
	}

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
