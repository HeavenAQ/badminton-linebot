package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/HeavenAQ/api/db"
	"github.com/go-resty/resty/v2"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/mowshon/moviego"
	"google.golang.org/api/drive/v3"
)

type AnalyzedResult struct {
	SkeletonVideo string   `json:"skeleton_video"`
	Suggestions   []string `json:"suggestions"`
	Score         string   `json:"score"`
}

func downloadVideo(app App, event *linebot.Event) (io.Reader, error) {
	app.InfoLogger.Println("\n\tDownloading video:")
	blob, err := app.downloadVideo(event)
	if err != nil {
		return nil, err
	}
	return blob, nil
}

func uploadVideoToDrive(app App, user *db.UserData, session *db.UserSession, skeleton_video string) (*drive.File, error) {
	app.InfoLogger.Println("\n\tUploading video:")
	folderId := app.getVideoFolder(user, session.Skill)
	driveFile, err := app.Drive.UploadVideo(folderId, skeleton_video)
	if err != nil {
		return nil, err
	}
	return driveFile, nil
}

func updateUserPortfolioVideo(app App, user *db.UserData, session *db.UserSession, driveFile *drive.File, aiRating string, aiSuggestions []string) error {
	app.InfoLogger.Println("\n\tUpdating user portfolio:")
	userPortfolio := app.getUserPortfolio(user, session.Skill)
	rating, err := strconv.ParseFloat(aiRating, 32)
	if err != nil {
		return err
	}

	for i, suggestion := range aiSuggestions {
		aiSuggestions[i] = fmt.Sprintf("%d. %s", i+1, suggestion)
	}

	return app.Db.CreateUserPortfolioVideo(
		user,
		userPortfolio,
		session,
		driveFile,
		float32(rating),
		strings.Join(aiSuggestions, "\n"),
	)
}

func sendVideoUploadedReply(app App, event *linebot.Event, session *db.UserSession, user *db.UserData) error {
	app.InfoLogger.Println("\n\tVideo uploaded successfully.")
	_, err := app.Bot.SendVideoUploadedReply(
		event.ReplyToken,
		session.Skill,
		app.getVideoFolder(user, session.Skill),
	)

	return err
}
func resizeVideo(app App, blob io.Reader, user db.UserData) (string, error) {
	app.InfoLogger.Println("\n\tResizing video:")

	filename := user.Id + ".mp4"
	file, err := os.Create(filename)
	if err != nil {
		return "", errors.New("failed to create tmp file for resizing")
	}
	defer file.Close()
	io.Copy(file, blob)

	first, err := moviego.Load(filename)
	if err != nil {
		return "", errors.New("failed to load video for resizing")
	}

	err = first.Resize(1080, 1920).Output("resized" + filename).Run()
	if err != nil {
		return "", errors.New("failed to resize video")
	}

	go os.Remove(filename)
	return "resized" + filename, nil
}

func analyzeVideo(app App, blob io.Reader, user *db.UserData, session *db.UserSession) (*AnalyzedResult, error) {
	app.InfoLogger.Println("\n\tAnalyzing video:")
	// resize video to 1080 x 1920
	resized, err := resizeVideo(app, blob, *user)
	if err != nil {
		return nil, err
	}

	resizedBlob, err := os.ReadFile(resized)
	if err != nil {
		return nil, err
	}

	os.Remove(resized)

	// set up request body with video data
	date := time.Now().Format("2006-01-02-15-04")
	filename := user.Id + "_" + session.Skill + "_" + date + ".mp4"
	baseURL := "https://b3b2-140-117-176-208.ngrok-free.app/analyze"
	client := resty.New()
	client.SetTimeout(1 * time.Minute)
	resp, err := client.R().
		SetBasicAuth("admin", "thisisacomplicatedpassword").
		SetQueryParam("handedness", user.Handedness.String()).
		SetQueryParam("skill", session.Skill).
		SetFileReader("file", filename, bytes.NewReader(resizedBlob)).
		Post(baseURL)

	// parse response to json
	var result AnalyzedResult
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		app.ErrorLogger.Println("AI Server Response:\n" + string(resp.Body()))
		return nil, err
	}
	return &result, nil
}

func uploadError(app App, event *linebot.Event, err error, message string) {
	app.ErrorLogger.Println(message, err)
	app.Bot.SendDefaultErrorReply(event.ReplyToken)
}

func (app *App) resolveUploadVideo(event *linebot.Event, user *db.UserData, session *db.UserSession) {
	blob, err := downloadVideo(*app, event)
	if err != nil {
		uploadError(*app, event, err, "\n\tError downloading video:")
		return
	}

	// analyze video
	result, err := analyzeVideo(*app, blob, user, session)
	if err != nil {
		uploadError(*app, event, err, "\n\tError analyzing video:")
		return
	}

	// upload video to google drive
	driveFile, err := uploadVideoToDrive(*app, user, session, result.SkeletonVideo)
	if err != nil {
		uploadError(*app, event, err, "\n\tError uploading video:")
		return
	}

	// update user portfolio
	if err := updateUserPortfolioVideo(*app, user, session, driveFile, result.Score, result.Suggestions); err != nil {
		uploadError(*app, event, err, "\n\tError updating user portfolio:")
		return
	}

	// send video uploaded reply
	if err := sendVideoUploadedReply(*app, event, session, user); err != nil {
		uploadError(*app, event, err, "\n\tError sending video uploaded reply:")
		return
	}

	// reset user session
	app.resetUserSession(user.Id)
}
