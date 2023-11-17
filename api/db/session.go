package db

import (
	"os"

	"cloud.google.com/go/firestore"
)

func (handler *FirebaseHandler) GetSessionCollection() *firestore.CollectionRef {
	collection := os.Getenv("FIREBASE_SESSIONS")
	return handler.dbClient.Collection(collection)
}

func (handler *FirebaseHandler) GetUserSession(userId string) (*UserSession, error) {
	session, err := handler.GetSessionCollection().Doc(userId).Get(handler.ctx)
	if err != nil {
		return nil, err
	}
	var userSessioon UserSession
	session.DataTo(&userSessioon)
	return &userSessioon, nil
}

func (handler *FirebaseHandler) NewUserSession(userId string) (*UserSession, error) {
	newSession := UserSession{
		UserState: None,
		Skill:     "",
	}
	err := handler.UpdateUserSession(userId, newSession)
	if err != nil {
		return nil, err
	}
	return &newSession, nil
}

func (handler *FirebaseHandler) UpdateUserSession(userId string, userSession UserSession) error {
	_, err := handler.GetSessionCollection().Doc(userId).Set(handler.ctx, userSession)
	if err != nil {
		return err
	}
	return nil
}

func (handler *FirebaseHandler) UpdateSessionUserState(userId string, state UserState) error {
	userSession, err := handler.GetUserSession(userId)
	if err != nil {
		return err
	}
	userSession.UserState = state
	return handler.UpdateUserSession(userId, *userSession)
}

func (handler *FirebaseHandler) UpdateSessionUserSkill(userId string, skill string) error {
	userSession, err := handler.GetUserSession(userId)
	if err != nil {
		return err
	}
	userSession.Skill = skill
	return handler.UpdateUserSession(userId, *userSession)
}
