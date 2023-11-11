package db

import (
	"context"

	"cloud.google.com/go/firestore"
)

type FirebaseHandler struct {
	dbClient *firestore.Client
	ctx      context.Context
}

type UserMap map[string]UserData
type UserData struct {
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
