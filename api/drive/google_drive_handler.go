package drive

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"github.com/HeavenAQ/api/secret"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func NewGoogleDriveHandler() (*GoogleDriveHandler, error) {
	ctx := context.Background()

	// retrieve google drive credentials from secret manager
	secretName := secret.GetSecretNameString(os.Getenv("GOOGLE_DRIVE_CREDENTIALS"))
	driveCredentials, err := secret.AccessSecretVersion(secretName)
	if err != nil {
		return nil, err
	}

	// create google drive service
	srv, err := drive.NewService(ctx, option.WithCredentialsJSON(driveCredentials))
	if err != nil {
		return nil, err
	}

	return &GoogleDriveHandler{
		srv,
		os.Getenv("GOOGLE_DRIVE_ROOT_FOLDER_ID"),
	}, nil
}

func (handler *GoogleDriveHandler) CreateUserFolders(userId string, userName string) (*UserFolders, error) {
	folderNames := []string{
		userId,
		"lift",
		"drop",
		"netplay",
		"clear",
		"footwork",
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
		case "lift":
			userFolders.LiftFolderId = folder.Id
		case "drop":
			userFolders.DropFolderId = folder.Id
		case "netplay":
			userFolders.NetplayFolderId = folder.Id
		case "clear":
			userFolders.ClearFolderId = folder.Id
		case "footwork":
			userFolders.FootworkFolderId = folder.Id
		}
	}

	return &userFolders, nil
}

func (handler *GoogleDriveHandler) UploadVideo(folderId string, blob io.Reader, thumbnailPath string) (*drive.File, *drive.File, error) {
	// upload video first
	filename := time.Now().Format("2006-01-02-15-04")
	driveFile, err := handler.srv.Files.Create(&drive.File{
		Name:    filename,
		Parents: []string{folderId},
	}).Media(blob).Do()
	if err != nil {
		return nil, nil, err
	}

	// load thumbnail as bytes
	thumbnail, err := os.Open(thumbnailPath)
	if err != nil {
		return nil, nil, err
	}
	defer thumbnail.Close()
	thumbnailData := new(bytes.Buffer)
	_, err = io.Copy(thumbnailData, thumbnail)

	// upload thumbnail
	thumbnailFile, err := handler.srv.Files.Create(&drive.File{
		Name:    filename + "_thumbnail",
		Parents: []string{os.Getenv("GOOGLE_DRIVE_THUMBNAIL_FOLDER_ID")},
	}).Media(thumbnailData).Do()
	return driveFile, thumbnailFile, nil
}
