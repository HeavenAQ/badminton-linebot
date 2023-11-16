package app

import (
	"io"
	"net/http"
	"strings"

	"github.com/HeavenAQ/api/db"
	"github.com/HeavenAQ/api/line"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (app *App) createUser(userId string) db.UserData {
	username, err := app.Bot.GetUserName(userId)
	if err != nil {
		app.ErrorLogger.Println("Error getting new user's name:", err)
	}
	userFolders, err := app.Drive.CreateUserFolders(userId, username)
	if err != nil {
		app.ErrorLogger.Println("Error creating new user's folders:", err)
	}
	userData, err := app.Db.CreateUserData(userFolders)
	if err != nil {
		app.ErrorLogger.Println("Error creating new user's data:", err)
	}
	return userData[userId]
}

func (app *App) createUserIfNotExist(userId string) (user *db.UserData) {
	user, err := app.Db.GetUserData(userId)
	if err != nil {
		app.WarnLogger.Println("User not found, creating new user...")
		userData := app.createUser(userId)
		user = &userData
	}
	return
}

func (app *App) createUserSessionIfNotExist(userId string) (userSession *db.UserSession) {
	userSession, err := app.Db.GetUserSession(userId)
	if err != nil {
		app.WarnLogger.Println("Session not found, creating new session...")
		userSession, err = app.Db.NewUserSession(userId)
		if err != nil {
			app.ErrorLogger.Println("Error creating new session:", err)
		}
	}
	return
}

func (app *App) updateUserState(userId string, state db.UserState) {
	err := app.Db.UpdateSessionUserState(userId, state)
	if err != nil {
		app.WarnLogger.Println("Error updating user state:", err)
	}
}

func (app *App) updateSessionUserSkill(userId string, skill line.Skill) {
	err := app.Db.UpdateSessionUserSkill(userId, skill.String())
	if err != nil {
		app.WarnLogger.Println("Error updating user skill:", err)
	}
}

func (app *App) downloadVideo(event *linebot.Event) (io.Reader, error) {
	videoUrl := event.Message.(*linebot.VideoMessage).ContentProvider.OriginalContentURL
	resp, err := http.Get(videoUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the body into a byte slice
	blob, err := io.ReadAll(resp.Body)
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

func (app *App) handleVideoMessage(event *linebot.Event, user *db.UserData, session *db.UserSession) {
	replyToken := event.ReplyToken
	blob, err := app.downloadVideo(event)
	if err != nil {
		app.WarnLogger.Println("Error downloading video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	folderId := app.getVideoFolder(user, session.Skill)
	driveFile, err := app.Drive.UploadVideo(folderId, blob)
	userPortfolio := app.getUserPortfolio(user, session.Skill)
	if err != nil {
		app.WarnLogger.Println("Error uploading video:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}

	err = app.Db.UpdateUserPortfolio(user, userPortfolio, session, driveFile)
}
