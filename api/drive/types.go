package drive

import (
	"google.golang.org/api/drive/v3"
)

type GoogleDriveHandler struct {
	srv          *drive.Service
	RootFolderID string
}

type UserFolders struct {
	UserId        string
	UserName      string
	RootFolderId  string
	ServeFolderId string
	SmashFolderId string
	ClearFolderId string
}
