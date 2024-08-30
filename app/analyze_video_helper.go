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
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
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

	// if no suggestions, add a default one
	if aiSuggestions == nil || len(aiSuggestions) == 0 {
		aiSuggestions = []string{"動作標準，無須調整"}
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

func resizeVideo(app App, blob io.Reader, user *db.UserData) (string, error) {
	app.InfoLogger.Println("\n\tStart Resizing video:")

	filename := "/tmp/" + user.Id + ".mp4"
	file, err := os.Create(filename)
	if err != nil {
		return "", errors.New("failed to create tmp file for resizing")
	}
	defer file.Close()

	// Stream the video directly to disk to avoid memory duplication
	app.InfoLogger.Println("\n\tCopying video blob to disk")
	if _, err := io.Copy(file, blob); err != nil {
		return "", errors.New("failed to write video blob to disk")
	}

	// Use ffmpeg-go to resize the video
	app.InfoLogger.Println("\n\tResizing video")
	outputFilename := "/tmp/resized_" + user.Id + ".mp4"
	err = ffmpeg_go.Input(filename).
		Filter("scale", ffmpeg_go.Args{"1080:1920"}).
		Output(outputFilename, ffmpeg_go.KwArgs{
			"vsync":   "0",  // avoid audio sync issues
			"threads": "1",  // use 1 thread to avoid memory issues
			"b:v":     "1M", // set video bitrate to 1 Mbps
			"an":      "",   // remove audio
		}).
		Run()
	if err != nil {
		return "", errors.New("failed to resize video")
	}

	// Asynchronously remove the original file
	go func() {
		if err := os.Remove(filename); err != nil {
			app.InfoLogger.Println("Failed to remove temp file:", err)
		}
	}()

	app.InfoLogger.Println("\n\tVideo resized successfully.")
	return outputFilename, nil
}

func analyzeVideo(app App, resizedVideo string, user *db.UserData, session *db.UserSession) (*AnalyzedResult, error) {
	app.InfoLogger.Println("\n\tAnalyzing video:")
	// resize video to 1080 x 1920
	resizedBlob, err := os.ReadFile(resizedVideo)
	if err != nil {
		return nil, err
	}

	os.Remove(resizedVideo)

	// set up request body with video data
	app.InfoLogger.Println("\n\tSending video to AI server: " + os.Getenv("GENAI_URL"))
	date := time.Now().Format("2006-01-02-15-04")
	filename := user.Id + "_" + session.Skill + "_" + date + ".mp4"
	baseURL := os.Getenv("GENAI_URL") + "/analyze"
	client := resty.New()
	client.SetTimeout(1 * time.Minute)
	resp, err := client.R().
		SetBasicAuth(os.Getenv("GENAI_USER"), os.Getenv("GENAI_PASSWORD")).
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
		app.resetUserSession(user.Id)
		return
	}

	resizedVideoName, err := resizeVideo(*app, blob, user)
	if err != nil {
		uploadError(*app, event, err, "\n\tError resizing video:")
		app.resetUserSession(user.Id)
		return
	}

	// analyze video
	result, err := analyzeVideo(*app, resizedVideoName, user, session)
	if err != nil {
		uploadError(*app, event, err, "\n\tError analyzing video:")
		app.resetUserSession(user.Id)
		return
	}

	// upload video to google drive
	driveFile, err := uploadVideoToDrive(*app, user, session, result.SkeletonVideo)
	if err != nil {
		uploadError(*app, event, err, "\n\tError uploading video:")
		app.resetUserSession(user.Id)
		return
	}

	// update user portfolio
	if err := updateUserPortfolioVideo(*app, user, session, driveFile, result.Score, result.Suggestions); err != nil {
		uploadError(*app, event, err, "\n\tError updating user portfolio:")
		app.resetUserSession(user.Id)
		return
	}

	// wait the thumbnail to be generated
	app.InfoLogger.Println("\n\tWaiting for thumbnail...")
	err = app.Drive.WaitForThumbnail(driveFile.Id)
	if err != nil {
		app.WarnLogger.Println("\n\tError waiting for thumbnail:", err)
	}

	// send video uploaded reply
	if err := sendVideoUploadedReply(*app, event, session, user); err != nil {
		uploadError(*app, event, err, "\n\tError sending video uploaded reply:")
		app.resetUserSession(user.Id)
		return
	}

	// reset user session
	app.resetUserSession(user.Id)
}
