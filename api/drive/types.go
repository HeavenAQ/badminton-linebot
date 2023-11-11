package drive

import (
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

type GoogleDriveHandler struct {
	tokenSource  oauth2.TokenSource
	client       *http.Client
	srv          *drive.Service
	RootFolderID string
}

type UserFolders struct {
	UserId           string
	UserName         string
	RootFolderId     string
	LiftFolderId     string
	DropFolderId     string
	NetplayFolderId  string
	ClearFolderId    string
	FootworkFolderId string
}
