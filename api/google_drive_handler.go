package api

import (
	"context"
	"log"
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

func NewGoogleDriveHandler() *GoogleDriveHandler {
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
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	return &GoogleDriveHandler{
		tokenSource,
		client,
		srv,
		os.Getenv("GOOGLE_ROOT_FOLDER_ID"),
	}
}
