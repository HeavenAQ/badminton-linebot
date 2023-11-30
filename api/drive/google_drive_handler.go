package drive

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

func getConfigFromJSON() *oauth2.Config {
	b, err := os.ReadFile(os.Getenv("GOOGLE_CREDENTIALS"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}

func getClient(config *oauth2.Config, ctx *context.Context) *http.Client {
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	token, err := config.Exchange(*ctx, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}

	return config.Client(*ctx, token)
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
