package drive

import (
	"google.golang.org/api/drive/v3"
)

type GoogleDriveHandler struct {
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
