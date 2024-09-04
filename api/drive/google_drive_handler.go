package drive

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/HeavenAQ/api/secret"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func (handler *GoogleDriveHandler) WaitForThumbnail(fileId string) error {
	// Initial delay and max attempts for polling
	maxAttempts := 40

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Retrieve the file metadata
		file, err := handler.srv.Files.Get(fileId).Fields("thumbnailLink").Do()
		if err != nil {
			return err
		}

		// Check if the thumbnailLink is available
		if file.ThumbnailLink != "" {
			return nil
		}

		// Wait before retrying
		time.Sleep(time.Second)
	}

	return fmt.Errorf("thumbnail generation timed out for file ID %s", fileId)
}

func NewGoogleDriveHandler() (*GoogleDriveHandler, error) {
	ctx := context.Background()

	// get google credentials from secret manager
	secretName := secret.GetSecretNameString(os.Getenv("GOOGLE_DRIVE_CREDENTIALS"))
	googleDriveCredentials, err := secret.AccessSecretVersion(secretName)

	// init google drive service
	srv, err := drive.NewService(ctx, option.WithCredentialsJSON(googleDriveCredentials))
	if err != nil {
		return nil, err
	}

	return &GoogleDriveHandler{
		srv, os.Getenv("GOOGLE_DRIVE_ROOT_FOLDER_ID"),
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

func (handler *GoogleDriveHandler) UploadVideo(folderId string, videoBlob []byte, thumbnailPath string) (*drive.File, *drive.File, error) {
	filename := time.Now().Format("2006-01-02-15-04")

	// upload video file to google drive
	blob := bytes.NewReader(videoBlob)
	driveFile, err := handler.srv.Files.Create(&drive.File{
		Name:    filename,
		Parents: []string{folderId},
	}).Media(blob).Do()
	if err != nil {
		return nil, nil, err
	}

	// upload video thumbnail to google drive
	thumbnailData, err := os.ReadFile(thumbnailPath)
	thumbnailFile, err := handler.srv.Files.Create(&drive.File{
		Name:    filename + "_thumbnail",
		Parents: []string{os.Getenv("GOOGLE_DRIVE_THUMBNAIL_FOLDER_ID")},
	}).Media(bytes.NewReader(thumbnailData)).Do()
	if err != nil {
		return nil, nil, err
	}

	return driveFile, thumbnailFile, nil
}
