package drive

import (
	"context"
	"io"
	"os"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func NewGoogleDriveHandler() (*GoogleDriveHandler, error) {
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_CREDENTIALS")))
	if err != nil {
		return nil, err
	}

	return &GoogleDriveHandler{
		srv,
		os.Getenv("GOOGLE_ROOT_FOLDER_ID"),
	}, nil
}

func (handler *GoogleDriveHandler) CreateUserFolders(userId string, userName string) (*UserFolders, error) {
	folderNames := []string{
		userId,
		"serve",
		"smash",
		"clear",
	}

	userFolders := UserFolders{
		UserId:   userId,
		UserName: userName,
	}

	for _, folderName := range folderNames {
		var parents []string
		if folderName == userId {
			parents = []string{handler.RootFolderID}
		} else {
			parents = []string{userFolders.RootFolderId}
		}

		folder, err := handler.srv.Files.Create(&drive.File{
			Name:     folderName,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  parents,
		}).Do()
		if err != nil {
			return nil, err
		}

		switch folderName {
		case userId:
			userFolders.RootFolderId = folder.Id
		case "serve":
			userFolders.ServeFolderId = folder.Id
		case "smash":
			userFolders.SmashFolderId = folder.Id
		case "clear":
			userFolders.ClearFolderId = folder.Id
		}
	}

	return &userFolders, nil
}

func (handler *GoogleDriveHandler) UploadVideo(folderId string, blob io.Reader) (*drive.File, error) {
	filename := time.Now().Format("2006-01-02-15-04")
	driveFile, err := handler.srv.Files.Create(&drive.File{
		Name:    filename,
		Parents: []string{folderId},
	}).Media(blob).Do()

	if err != nil {
		return nil, err
	}

	return driveFile, nil
}
