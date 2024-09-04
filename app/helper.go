package app

import (
	"bytes"
	"io"
	"os"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

func (app *App) rmTmpFile(filename string) {
	if err := os.Remove(filename); err != nil {
		app.InfoLogger.Println("Failed to remove temp file:", err)
	}
}

func (app *App) createUser(userId string) *db.UserData {
	username, err := app.Bot.GetUserName(userId)
	if err != nil {
		app.ErrorLogger.Println("\n\tError getting new user's name:", err)
	}
	userFolders, err := app.Drive.CreateUserFolders(userId, username)
	if err != nil {
		app.ErrorLogger.Println("\n\tError creating new user's folders:", err)
	}
	userData, err := app.Db.CreateUserData(userFolders)
	if err != nil {
		app.ErrorLogger.Println("\n\tError creating new user's data:", err)
	}
	return userData
}

func (app *App) createUserIfNotExist(userId string) (user *db.UserData) {
	user, err := app.Db.GetUserData(userId)
	if err != nil {
		app.WarnLogger.Println("\n\tUser not found, creating new user...")
		userData := app.createUser(userId)
		user = userData
		app.InfoLogger.Println("\n\tNew user created successfully.")
	}
	return
}

func (app *App) createUserSessionIfNotExist(userId string) (userSession *db.UserSession) {
	userSession, err := app.Db.GetUserSession(userId)
	if err != nil {
		app.WarnLogger.Println("\n\tSession not found, creating new session...")
		userSession, err = app.Db.NewUserSession(userId)
		if err != nil {
			app.ErrorLogger.Println("\n\tError creating new session:", err)
		} else {
			app.InfoLogger.Println("\n\tNew session created successfully.")
		}
	}
	return
}

func (app *App) updateUserState(userId string, state db.UserState) {
	err := app.Db.UpdateSessionUserState(userId, state)
	if err != nil {
		app.WarnLogger.Println("\n\tError updating user state:", err)
	}
}

func (app *App) resetUserSession(userId string) {
	err := app.Db.UpdateUserSession(
		userId,
		db.UserSession{
			UserState: db.None,
			Skill:     "",
		},
	)
	if err != nil {
		app.ErrorLogger.Println("\n\tError resetting user session:", err)
	}
}

func (app *App) downloadVideo(event *linebot.Event) (string, error) {
	resp, err := app.Bot.GetVideoContent(event)
	if err != nil {
		return "", err
	}
	defer resp.Content.Close()

	// Read the body into a file
	filename := "/tmp/" + event.WebhookEventID + ".mp4"
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(file, resp.Content)
	if err != nil {
		return "", err
	}
	resp.Content.Close()
	return filename, nil
}

func (app *App) getVideoFolder(user *db.UserData, skill string) string {
	var folderId string
	switch skill {
	case "lift":
		folderId = user.FolderIds.Lift
	case "drop":
		folderId = user.FolderIds.Drop
	case "netplay":
		folderId = user.FolderIds.Netplay
	case "clear":
		folderId = user.FolderIds.Clear
	case "footwork":
		folderId = user.FolderIds.Footwork
	}
	return folderId
}

func (app *App) getUserPortfolio(user *db.UserData, skill string) *map[string]db.Work {
	var work map[string]db.Work
	switch skill {
	case "lift":
		work = user.Portfolio.Lift
	case "drop":
		work = user.Portfolio.Drop
	case "netplay":
		work = user.Portfolio.Netplay
	case "clear":
		work = user.Portfolio.Clear
	case "footwork":
		work = user.Portfolio.Footwork
	}
	return &work
}

func (app *App) updateUserReflection(event *linebot.Event, user *db.UserData, session *db.UserSession) error {
	userPortfolio := app.getUserPortfolio(user, session.Skill)
	reflection := event.Message.(*linebot.TextMessage).Text
	err := app.Db.UpdateUserPortfolioReflection(user, userPortfolio, session, reflection)
	if err != nil {
		return err
	}
	_, err = app.Bot.SendReflectionUpdatedReply(event.ReplyToken)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) createVideoThumbnail(event *linebot.Event, user *db.UserData, videoFilePath string) (string, error) {
	// Using ffmpeg to create video thumbnail
	app.InfoLogger.Println("Extracting thumbnail from the video")
	outFileName := "/tmp/" + user.Id + ".jpeg"
	err := ffmpeg_go.Input(videoFilePath).
		Output(outFileName, ffmpeg_go.KwArgs{
			"vframes": 1,                        // extract exactly 1 frame
			"vcodec":  "mjpeg",                  // make it a jpeg file
			"vf":      "thumbnail,scale=320:-1", // scale the image to 320px width, keep aspect ratio
			"ss":      "00:00:01",               // extract frame at 1 second
		}).
		Run()
	if err != nil {
		app.ErrorLogger.Println("\n\tError extracting thumbnail from video:", err)
		return "", err
	}

	return outFileName, nil
}

func (app *App) handleVideoMessage(event *linebot.Event, user *db.UserData, session *db.UserSession) {
	replyToken := event.ReplyToken

	// download video from line
	app.InfoLogger.Println("\n\tDownloading video...")
	video, err := app.downloadVideo(event)
	if err != nil {
		app.WarnLogger.Println("\n\tError downloading video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	// create thumbnail from the blob
	app.InfoLogger.Println("\n\tCreating thumbnail...")
	thumbnail, err := app.createVideoThumbnail(event, user, video)
	if err != nil {
		app.ErrorLogger.Println("\n\tError creating thumbnail")
	}

	// open video videoData to upload
	videoData, err := os.ReadFile(video)
	if err != nil {
		app.WarnLogger.Println("\n\tFailed to open video file")
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	// upload video to google drive
	app.InfoLogger.Println("\n\tUploading video:")
	folderId := app.getVideoFolder(user, session.Skill)
	driveFile, thumbnailFile, err := app.Drive.UploadVideo(folderId, bytes.NewReader(videoData), thumbnail)
	if err != nil {
		app.WarnLogger.Println("\n\tError uploading video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	// update user portfolio
	userPortfolio := app.getUserPortfolio(user, session.Skill)
	err = app.Db.CreateUserPortfolioVideo(user, userPortfolio, session, driveFile, thumbnailFile)
	if err != nil {
		app.WarnLogger.Println("\n\tError updating user portfolio after uploading video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	// Asynchronously remove the thumbnail file
	go app.rmTmpFile(thumbnail)
	go app.rmTmpFile(video)

	// send video uploaded reply
	app.InfoLogger.Println("\n\tVideo uploaded successfully.")
	_, err = app.Bot.SendVideoUploadedReply(
		event.ReplyToken,
		session.Skill,
		app.getVideoFolder(user, session.Skill),
	)
	if err != nil {
		app.WarnLogger.Println("\n\tError sending video uploaded reply:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	// reset user session
	app.resetUserSession(user.Id)
}
