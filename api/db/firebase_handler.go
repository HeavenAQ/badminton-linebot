package db

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/HeavenAQ/api/drive"
)

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
	return &FirebaseHandler{client, ctx}, nil
}

func (handler *FirebaseHandler) GetCollection() *firestore.CollectionRef {
	collection := os.Getenv("FIREBASE_COLLECTION")
	return handler.dbClient.Collection(collection)
}

func (handler *FirebaseHandler) CreateUserData(userFolders *drive.UserFolders) (UserMap, error) {
	ref := handler.GetCollection().NewDoc()
	newUserTemplate := UserMap{
		userFolders.UserId: {
			Name:       userFolders.UserName,
			Id:         userFolders.UserId,
			Handedness: Right,
			FolderIds: FolderIds{
				Root:     userFolders.RootFolderId,
				Lift:     userFolders.LiftFolderId,
				Drop:     userFolders.DropFolderId,
				Netplay:  userFolders.NetplayFolderId,
				Clear:    userFolders.ClearFolderId,
				Footwork: userFolders.FootworkFolderId,
			},
			Portfolio: Portfolio{
				Lift:     map[string]Work{},
				Drop:     map[string]Work{},
				Netplay:  map[string]Work{},
				Clear:    map[string]Work{},
				Footwork: map[string]Work{},
			},
		},
	}

	_, err := ref.Set(handler.ctx, newUserTemplate)
	if err != nil {
		return nil, err
	}
	return newUserTemplate, nil
}

func (handler *FirebaseHandler) GetUserData(userId string) (*UserData, error) {
	docsnap, err := handler.GetCollection().Doc(userId).Get(handler.ctx)
	if err != nil {
		return nil, err
	}
	var user UserData
	docsnap.DataTo(&user)
	return &user, nil
}

func (handler *FirebaseHandler) updateUserData(user *UserData) error {
	_, err := handler.GetCollection().Doc(user.Id).Set(handler.ctx, UserMap{user.Id: *user})
	if err != nil {
		return err
	}
	return nil
}

func (handler *FirebaseHandler) UpdateUserHandedness(user *UserData, handedness Handedness) error {
	user.Handedness = handedness
	return handler.updateUserData(user)
}
