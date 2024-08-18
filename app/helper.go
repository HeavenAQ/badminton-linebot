package app

import (
	"io"
	"strings"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

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

func (app *App) downloadVideo(event *linebot.Event) (io.Reader, error) {
	app.InfoLogger.Println("\n\tDownloading video...")
	resp, err := app.Bot.GetVideoContent(event)
	if err != nil {
		return nil, err
	}
	defer resp.Content.Close()

	// Read the body into a byte slice
	blob, err := io.ReadAll(resp.Content)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(string(blob)), nil
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

func (app *App) handleVideoMessage(event *linebot.Event, user *db.UserData, session *db.UserSession) {
	replyToken := event.ReplyToken
	app.InfoLogger.Println("\n\tDownloading video:")

	// download video from line
	blob, err := app.downloadVideo(event)
	if err != nil {
		app.WarnLogger.Println("\n\tError downloading video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	// upload video to google drive
	app.InfoLogger.Println("\n\tUploading video:")
	folderId := app.getVideoFolder(user, session.Skill)
	driveFile, err := app.Drive.UploadVideo(folderId, blob)
	if err != nil {
		app.WarnLogger.Println("\n\tError uploading video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	// Poll for the thumbnail to be ready
	app.InfoLogger.Println("\n\tWaiting for thumbnail...")
	err = app.Drive.WaitForThumbnail(driveFile.Id)
	if err != nil {
		app.WarnLogger.Println("\n\tError waiting for thumbnail:", err)
	}

	// update user portfolio
	userPortfolio := app.getUserPortfolio(user, session.Skill)
	err = app.Db.CreateUserPortfolioVideo(user, userPortfolio, session, driveFile)
	if err != nil {
		app.WarnLogger.Println("\n\tError updating user portfolio after uploading video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

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
