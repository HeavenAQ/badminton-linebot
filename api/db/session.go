package db

import (
	"os"

	"cloud.google.com/go/firestore"
)

func (handler *FirebaseHandler) GetSessionCollection() *firestore.CollectionRef {
	collection := os.Getenv("FIREBASE_SESSION")
	return handler.dbClient.Collection(collection)
}

func (handler *FirebaseHandler) GetSession(userId string) (*UserSession, error) {
	session, err := handler.GetSessionCollection().Doc(userId).Get(handler.ctx)
	if err != nil {
		return nil, err
	}
	var userSessioon UserSession
	session.DataTo(&userSessioon)
	return &userSessioon, nil
}

func (handler *FirebaseHandler) UpdateSession(userId string, newSession UserSession) error {
	_, err := handler.GetSessionCollection().Doc(userId).Set(handler.ctx, newSession)
	if err != nil {
		return err
	}
	return nil
}
