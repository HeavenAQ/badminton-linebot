package app

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/HeavenAQ/api/db"
	"github.com/HeavenAQ/api/drive"
	"github.com/HeavenAQ/api/line"
	"github.com/alexedwards/scs/v2"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type App struct {
	Bot         *line.LineBotHandler
	Drive       *drive.GoogleDriveHandler
	Db          *db.FirebaseHandler
	RootFolder  string
	Session     *scs.SessionManager
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	WarnLogger  *log.Logger
}

func NewApp() *App {
	rootFolder := os.Getenv("GOOGLE_ROOT_FOLDER_ID")
	infoLogger := log.New(log.Writer(), "[INFO] ", log.LstdFlags)
	errorLogger := log.New(log.Writer(), "[ERROR] ", log.LstdFlags)
	warnLogger := log.New(log.Writer(), "[WARN] ", log.LstdFlags)

	db, err := db.NewFirebaseHandler()
	if err != nil {
		errorLogger.Println("\n\tError initializing firebase database client:", err)
	}

	bot, err := line.NewLineBotHandler()
	if err != nil {
		errorLogger.Println("\n\tError initializing line bot client:", err)
	}

	drive, err := drive.NewGoogleDriveHandler()
	if err != nil {
		errorLogger.Println("\n\tError initializing google drive client:", err)
	}

	infoLogger.Println("\n\tApp initialized successfully.")
	return &App{
		Bot:         bot,
		Drive:       drive,
		Db:          db,
		RootFolder:  rootFolder,
		InfoLogger:  infoLogger,
		ErrorLogger: errorLogger,
		WarnLogger:  warnLogger,
	}
}

func (app *App) HandleCallback(w http.ResponseWriter, req *http.Request) {
	// retrieve events
	events, err := app.Bot.RetrieveCbEvent(w, req)
	if err != nil {
		app.ErrorLogger.Println("\n\tError retrieving callback event:", err)
		return
	}

	// handle events
	for _, event := range events {
		// get user
		user := app.createUserIfNotExist(event.Source.UserID)
		session := app.createUserSessionIfNotExist(event.Source.UserID)
		app.InfoLogger.Println(
			"\n\tIncoming event:", event.Type,
			"\n\t\t- User (", user.Id, ")",
			"\n\t\t- Session: ", session,
		)

		// handler event
		switch event.Type {
		case linebot.EventTypeFollow:
			app.Bot.SendWelcomeReply(event)
		case linebot.EventTypeMessage:
			app.handleMessageEvent(event, user, session)
		case linebot.EventTypePostback:
			app.handlePostbackEvent(event, user)
		default:
			app.WarnLogger.Println("\n\tUnknown event type: ", event.Type)
			app.Bot.SendDefaultReply(event.ReplyToken)

		}
	}
}

func (app *App) handleMessageEvent(event *linebot.Event, user *db.UserData, session *db.UserSession) {
	switch event.Message.(type) {
	case *linebot.TextMessage:
		app.handleTextMessage(event, user, session)
	case *linebot.VideoMessage:
		if session.UserState == db.UploadingVideo {
			app.handleVideoMessage(event, user, session)
		} else {
			app.WarnLogger.Println("\n\tUnknown message type: ", event.Message.Type())
			app.Bot.SendDefaultErrorReply(event.ReplyToken)
		}
	default:
		app.WarnLogger.Println("\n\tUnknown message type: ", event.Message.Type())
		app.Bot.SendDefaultReply(event.ReplyToken)
	}
}

func (app *App) handleTextMessage(event *linebot.Event, user *db.UserData, session *db.UserSession) {
	replyToken := event.ReplyToken
	switch event.Message.(*linebot.TextMessage).Text {
	case "使用說明":
		res, err := app.Bot.SendInstruction(replyToken)
		if err != nil {
			app.WarnLogger.Println("\n\tError sending instruction: ", err)
		}
		app.InfoLogger.Println("\n\tInstruction sent. Response from line: ", res)
	case "我的學習歷程":
		app.Bot.PromptSkillSelection(replyToken, line.ViewPortfolio, "請選擇要查看的學習歷程")
	case "專家影片":
		_, err := app.Bot.PromptHandednessSelection(replyToken)
		if err != nil {
			app.ErrorLogger.Println("\n\tError prompting handedness selection: ", err)
		}
	case "上傳錄影":
		app.Bot.PromptSkillSelection(replyToken, line.Upload, "請選擇要上傳錄影的動作")
		go app.updateUserState(user.Id, db.UploadingVideo)
	case "新增學習反思":
		app.Bot.PromptSkillSelection(replyToken, line.AddReflection, "請選擇要新增學習反思的動作")
		go app.updateUserState(user.Id, db.WritingReflection)
	case "課程大綱":
		app.Bot.SendSyllabus(replyToken)
	default:
		isWritingReflection := session.UserState == db.WritingReflection
		if isWritingReflection {
			err := app.updateUserReflection(event, user, session)
			if err != nil {
				app.WarnLogger.Println("\n\tError updating user reflection:", err)
				app.Bot.SendDefaultErrorReply(replyToken)
			}
			app.resetUserSession(user.Id)
		} else {
			app.Bot.SendDefaultReply(replyToken)
		}
	}
}

func (app *App) handlePostbackEvent(event *linebot.Event, user *db.UserData) {
	app.InfoLogger.Println("\n\tPostback event:", event.Postback.Data)
	replyToken := event.ReplyToken
	tmp := strings.Split(event.Postback.Data, "&")
	var data [2][2]string
	for i, t := range tmp {
		data[i] = [2]string(strings.SplitN(t, "=", 2))
	}

	if len(data) == 0 {
		app.WarnLogger.Println("\n\tEmpty postback data")
	} else if data[0][0] == "handedness" {
		app.handleHandednessReply(replyToken, user, data[0][1])
	} else {
		app.handleUserAction(event, user, data)
	}
}

func (app *App) handleHandednessReply(replyToken string, user *db.UserData, data string) {
	handedness, err := db.HandednessStrToEnum(data)
	if err != nil {
		app.WarnLogger.Println("\n\tInvalid handedness data")
		app.Bot.SendWrongHandednessReply(replyToken)
		return
	}

	if user.Handedness != handedness {
		err = app.Db.UpdateUserHandedness(user, handedness)
		if err != nil {
			app.WarnLogger.Println("\n\tError updating user handedness:", err)
			app.Bot.SendDefaultErrorReply(replyToken)
			return
		}
	}
	app.Bot.PromptSkillSelection(replyToken, line.ViewExpertVideo, "請選擇要觀看的動作")
}

func (app *App) handleUserAction(event *linebot.Event, user *db.UserData, data [2][2]string) {
	replyToken := event.ReplyToken
	var userAction line.UserActionPostback
	err := userAction.FromArray(data)
	if err != nil {
		app.WarnLogger.Println("\n\tInvalid postback data")
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}
	err = app.ResolveUserAction(event, user, userAction)
	if err != nil {
		app.ErrorLogger.Println("\n\tError resolving user action:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}
}

func (app *App) ResolveUserAction(event *linebot.Event, user *db.UserData, action line.UserActionPostback) error {
	switch action.Type {
	case line.AddReflection:
		// update user session
		go app.Db.UpdateUserSession(user.Id, db.UserSession{
			UserState: db.WritingReflection,
			Skill:     action.Skill.String(),
		})

		app.Bot.ResolveAddReflection(event, user, action.Skill)
	case line.ViewPortfolio:
		err := app.Bot.ResolveViewPortfolio(event, user, action.Skill)
		if err != nil {
			return errors.New("\n\tError resolving view portfolio: " + err.Error())
		}
	case line.ViewExpertVideo:
		err := app.Bot.ResolveViewExpertVideo(event, user, action.Skill)
		if err != nil {
			return errors.New("\n\tError resolving view expert video: " + err.Error())
		}
	case line.Upload:
		// update user session
		go app.Db.UpdateUserSession(user.Id, db.UserSession{
			UserState: db.UploadingVideo,
			Skill:     action.Skill.String(),
		})

		err := app.Bot.ResolveVideoUpload(event, user, action.Skill)
		if err != nil {
			return errors.New("\n\tError resolving upload: " + err.Error())
		}
	default:
		return errors.New("\n\tInvalid user action type")
	}
	return nil
}
