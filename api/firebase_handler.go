package api

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

type FirebaseHandler struct {
	dbClient *firestore.Client
}

func NewFirebaseHandler() (*FirebaseHandler, error) {
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: os.Getenv("FIREBASE_PROJECT_ID")}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatal("Error initializing firebase database client:", err)
	}
	return &FirebaseHandler{client}, nil
}
