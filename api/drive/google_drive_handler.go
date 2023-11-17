package drive

import (
	"context"
	"io"
	"os"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

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
