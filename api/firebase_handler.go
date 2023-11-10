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
	ctx      context.Context
}

type UserTemplate struct {
	Name      string    `json:"name"`
	FolderIds FolderIds `json:"folderIds"`
	Portfolio Portfolio `json:"portfolio"`
}

type FolderIds struct {
	Root     string `json:"root"`
	Lift     string `json:"lift"`
	Drop     string `json:"drop"`
	Netplay  string `json:"netplay"`
	Clear    string `json:"clear"`
	Footwork string `json:"footwork"`
}

type Portfolio struct {
	Lift     map[string]map[string]Work `json:"lift"`
	Drop     map[string]map[string]Work `json:"drop"`
	Netplay  map[string]map[string]Work `json:"netplay"`
	Clear    map[string]map[string]Work `json:"clear"`
	Footwork map[string]map[string]Work `json:"footwork"`
}

type Work struct {
	Date       string `json:"date"`
	Video      string `json:"video"`
	Thumbnail  string `json:"thumbnail"`
	Reflection string `json:"reflection"`
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
	return &FirebaseHandler{client, ctx}, nil
}

func (handler *FirebaseHandler) GetCollection() *firestore.CollectionRef {
	collection := os.Getenv("FIREBASE_COLLECTION")
	return handler.dbClient.Collection(collection)
}

func (handler *FirebaseHandler) CreateUserData(userFolders *UserFolders) (map[string]UserTemplate, error) {
	ref := handler.GetCollection().NewDoc()
	newUserTemplate := map[string]UserTemplate{
		userFolders.userId: {
			Name: userFolders.userName,
			FolderIds: FolderIds{
				Root:     userFolders.rootFolderId,
				Lift:     userFolders.liftFolderId,
				Drop:     userFolders.dropFolderId,
				Netplay:  userFolders.netplayFolderId,
				Clear:    userFolders.clearFolderId,
				Footwork: userFolders.footworkFolderId,
			},
			Portfolio: Portfolio{
				Lift:     map[string]map[string]Work{},
				Drop:     map[string]map[string]Work{},
				Netplay:  map[string]map[string]Work{},
				Clear:    map[string]map[string]Work{},
				Footwork: map[string]map[string]Work{},
			},
		},
	}

	_, err := ref.Set(handler.ctx, newUserTemplate)
	if err != nil {
		return nil, err
	}
	return newUserTemplate, nil
}

func (handler *FirebaseHandler) GetUserData(userId string) (map[string]UserTemplate, error) {
	docsnap, err := handler.GetCollection().Doc(userId).Get(handler.ctx)
	if err != nil {
		return nil, err
	}
	var userData map[string]UserTemplate
	docsnap.DataTo(&userData)
	return userData, nil
}

func (handler *FirebaseHandler) Close() {
	handler.dbClient.Close()
}
