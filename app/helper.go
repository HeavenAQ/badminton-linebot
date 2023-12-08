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
	case "serve":
		folderId = user.FolderIds.Serve
	case "smash":
		folderId = user.FolderIds.Smash
	case "clear":
		folderId = user.FolderIds.Clear
	}
	return folderId
}

func (app *App) getUserPortfolio(user *db.UserData, skill string) *map[string]db.Work {
	var work map[string]db.Work
	switch skill {
	case "serve":
		work = user.Portfolio.Serve
	case "smash":
		work = user.Portfolio.Smash
	case "clear":
		work = user.Portfolio.Clear
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
	_, err = app.Bot.SendReply(event.ReplyToken, "已成功更新個人學習反思!")
	if err != nil {
		return err
	}
	return nil
}

func (app *App) updateUserPreviewNote(event *linebot.Event, user *db.UserData, session *db.UserSession) error {
	userPortfolio := app.getUserPortfolio(user, session.Skill)
	previewNote := event.Message.(*linebot.TextMessage).Text
	err := app.Db.UpdateUserPortfolioPreviewNote(user, userPortfolio, session, previewNote)
	if err != nil {
		return err
	}
	_, err = app.Bot.SendReply(event.ReplyToken, "已成功更新課前檢視要點!")
	if err != nil {
		return err
	}
	return nil
}

func (app *App) resolveWritingReflection(event *linebot.Event, user *db.UserData, session *db.UserSession) error {
	switch event.Message.(type) {
	case *linebot.TextMessage:
		err := app.updateUserReflection(event, user, session)
		if err != nil {
			return err
		}
		app.resetUserSession(user.Id)
	default:
		_, err := app.Bot.SendReply(event.ReplyToken, "請輸入學習反思")
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *App) resolveWritingPreviewNote(event *linebot.Event, user *db.UserData, session *db.UserSession) error {
	switch event.Message.(type) {
	case *linebot.TextMessage:
		err := app.updateUserPreviewNote(event, user, session)
		if err != nil {
			return err
		}
		app.resetUserSession(user.Id)
	default:
		_, err := app.Bot.SendReply(event.ReplyToken, "請輸入課前檢視要點")
		if err != nil {
			return err
		}
	}
	return nil
}
