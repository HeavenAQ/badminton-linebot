package api

import (
	"context"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleDriveHandler struct {
	tokenSource  oauth2.TokenSource
	client       *http.Client
	srv          *drive.Service
	RootFolderID string
}

type UserFolders struct {
	userId           string
	userName         string
	rootFolderId     string
	liftFolderId     string
	dropFolderId     string
	netplayFolderId  string
	clearFolderId    string
	footworkFolderId string
}

func NewGoogleDriveHandler() (*GoogleDriveHandler, error) {
	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{drive.DriveScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}

	tokenSource := config.TokenSource(ctx, &oauth2.Token{RefreshToken: os.Getenv("GOOGLE_REFRESH_TOKEN")})
	client := oauth2.NewClient(ctx, tokenSource)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &GoogleDriveHandler{
		tokenSource,
		client,
		srv,
		os.Getenv("GOOGLE_ROOT_FOLDER_ID"),
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
		userId:   userId,
		userName: userName,
	}

	for _, folderName := range folderNames {
		folder, err := handler.srv.Files.Create(&drive.File{
			Name:     folderName,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{handler.RootFolderID},
		}).Do()
		if err != nil {
			return nil, err
		}

		switch folderName {
		case userId:
			userFolders.rootFolderId = folder.Id
		case "lift":
			userFolders.liftFolderId = folder.Id
		case "drop":
			userFolders.dropFolderId = folder.Id
		case "netplay":
			userFolders.netplayFolderId = folder.Id
		case "clear":
			userFolders.clearFolderId = folder.Id
		case "footwork":
			userFolders.footworkFolderId = folder.Id
		}
	}

	return &userFolders, nil
}
