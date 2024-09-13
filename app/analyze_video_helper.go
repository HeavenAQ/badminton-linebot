package app

import (
	"bytes"
	"encoding/base64"
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
	Score         string   `json:"score"`
	Suggestions   []string `json:"suggestions"`
}

func downloadVideo(app App, event *linebot.Event) (io.Reader, error) {
	app.InfoLogger.Println("\n\tDownloading video:")
	blob, err := app.downloadVideo(event)
	if err != nil {
		return nil, err
	}
	return blob, nil
}

func uploadVideoToDrive(app App, user *db.UserData, session *db.UserSession, skeletonVideo []byte, thumbnailPath string) (*drive.File, *drive.File, error) {
	app.InfoLogger.Println("\n\tUploading video:")
	folderId := app.getVideoFolder(user, session.Skill)
	driveFile, thumbnailFile, err := app.Drive.UploadVideo(folderId, skeletonVideo, thumbnailPath)
	if err != nil {
		return nil, nil, err
	}
	return driveFile, thumbnailFile, nil
}

func updateUserPortfolioVideo(app App, user *db.UserData, session *db.UserSession, driveFile *drive.File, thumbnailFile *drive.File, aiRating string, aiSuggestions []string) error {
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
	if len(aiSuggestions) == 0 {
		aiSuggestions = []string{"動作標準，無須調整"}
	}

	return app.Db.CreateUserPortfolioVideo(
		user,
		userPortfolio,
		session,
		driveFile,
		thumbnailFile,
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

func createTmpVideoFile(app App, blob io.Reader, user *db.UserData) (string, error) {
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

	return filename, nil
}

func rmTmpVideoFile(app App, filename string) {
	app.InfoLogger.Println("\n\tRemoving tmp video file")
	if err := os.Remove(filename); err != nil {
		app.WarnLogger.Println("Failed to remove tmp video file:", err)
	}
}

func resizeVideo(app App, user *db.UserData, videoPath string) (string, error) {
	// Use ffmpeg-go to resize the video
	app.InfoLogger.Println("\n\tStart Resizing video:")
	app.InfoLogger.Println("\n\tResizing video")
	outputFilename := "/tmp/resized_" + user.Id + ".mp4"
	err := ffmpeg_go.Input(videoPath).
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

	maxRetries := 3
	delay := 5 * time.Second
	var resp *resty.Response
	for i := 0; i < maxRetries; i++ {
		// send video to AI server
		resp, err = client.R().
			SetBasicAuth(os.Getenv("GENAI_USER"), os.Getenv("GENAI_PASSWORD")).
			SetQueryParam("handedness", user.Handedness.String()).
			SetQueryParam("skill", session.Skill).
			SetFileReader("file", filename, bytes.NewReader(resizedBlob)).
			Post(baseURL)

		// if no error and status code is not 502, break the loop
		if err == nil && resp.StatusCode() != 502 {
			break
			// if status code is 502, retry after 5 seconds
		} else if resp != nil && resp.StatusCode() == 502 {
			app.WarnLogger.Println("AI Server is busy, retrying in 5 seconds")
			time.Sleep(delay)
			// if error is not nil, return error
		} else if err != nil {
			return nil, err
		}
	}

	// Check if we have a valid response
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("failed to get response from AI server after %d retries", maxRetries)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code %d from AI server", resp.StatusCode())
	}

	// parse response to json
	var result AnalyzedResult
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		app.ErrorLogger.Println("AI Server Response:\n" + string(resp.Body()))
		return nil, err
	}
	return &result, nil
}

func (app *App) createVideoThumbnail(event *linebot.Event, user *db.UserData, blob []byte) (string, error) {
	// create a tmp file to store video blob
	app.InfoLogger.Println("\n\tCreating a tmp file to store video blob ...")
	replyToken := event.ReplyToken
	filename := "/tmp/" + user.Id + ".mp4"
	file, err := os.Create(filename)
	if err != nil {
		app.ErrorLogger.Println("\n\tError creating tmp file for video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return "", err
	}
	defer file.Close()

	// write video blob to the tmp file
	app.InfoLogger.Println("\n\tWriting video blob to tmp file")
	if _, err := io.Copy(file, bytes.NewReader(blob)); err != nil {
		app.ErrorLogger.Println("\n\tError writing video blob to tmp file:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return "", err
	}

	// Using ffmpeg to create video thumbnail
	app.InfoLogger.Println("Extracting thumbnail from the video")
	outFileName := "/tmp/" + user.Id + ".jpeg"

	var stderr bytes.Buffer
	err = ffmpeg_go.Input(filename, ffmpeg_go.KwArgs{
		"ss": "00:00:01", // place ss before input file to avoid seeking issues
	}).
		Output(outFileName, ffmpeg_go.KwArgs{
			"vframes": 1,              // extract exactly 1 frame
			"vcodec":  "mjpeg",        // make it a jpeg file
			"vf":      "scale=320:-1", // scale the image to 320px width, keep aspect ratio
		}).
		WithErrorOutput(&stderr). // Capture stderr for debugging
		Run()
	if err != nil {
		app.ErrorLogger.Println("\n\tError extracting thumbnail from video:", err)
		app.ErrorLogger.Println("\n\tffmpeg stderr:", stderr.String())
		return "", err
	}

	// Asynchronously remove the original file
	go func() {
		if err := os.Remove(filename); err != nil {
			app.InfoLogger.Println("Failed to remove temp file:", err)
		}
	}()
	return outFileName, nil
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

	// create tmp video file
	videoPath, err := createTmpVideoFile(*app, blob, user)
	if err != nil {
		uploadError(*app, event, err, "\n\tError creating tmp video file:")
		app.resetUserSession(user.Id)
		return
	}

	resizedVideoPath, err := resizeVideo(*app, user, videoPath)
	if err != nil {
		uploadError(*app, event, err, "\n\tError resizing video:")
		app.resetUserSession(user.Id)
		return
	}

	// analyze video
	result, err := analyzeVideo(*app, resizedVideoPath, user, session)
	if err != nil {
		uploadError(*app, event, err, "\n\tError analyzing video:")
		app.resetUserSession(user.Id)
		return
	}

	// decode skeleton video
	decodedVideo, err := base64.StdEncoding.DecodeString(result.SkeletonVideo)
	if err != nil {
		uploadError(*app, event, err, "\n\tError decoding video:")
		app.resetUserSession(user.Id)
		return
	}

	// create video thumbnail
	thumbnailPath, err := app.createVideoThumbnail(event, user, decodedVideo)
	if err != nil {
		uploadError(*app, event, err, "\n\tError creating video thumbnail:")
		app.resetUserSession(user.Id)
		return
	}

	// upload video to google drive
	driveFile, thumbnailFile, err := uploadVideoToDrive(*app, user, session, decodedVideo, thumbnailPath)
	if err != nil {
		uploadError(*app, event, err, "\n\tError uploading video:")
		app.resetUserSession(user.Id)
		return
	}

	// update user portfolio
	if err := updateUserPortfolioVideo(*app, user, session, driveFile, thumbnailFile, result.Score, result.Suggestions); err != nil {
		uploadError(*app, event, err, "\n\tError updating user portfolio:")
		app.resetUserSession(user.Id)
		return
	}

	// send video uploaded reply
	if err := sendVideoUploadedReply(*app, event, session, user); err != nil {
		uploadError(*app, event, err, "\n\tError sending video uploaded reply:")
		app.resetUserSession(user.Id)
		return
	}

	// reset user session
	rmTmpVideoFile(*app, videoPath)
	rmTmpVideoFile(*app, resizedVideoPath)
	rmTmpVideoFile(*app, thumbnailPath)
	app.resetUserSession(user.Id)
}
