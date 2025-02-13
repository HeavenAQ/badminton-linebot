package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
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
	Session     *scs.SessionManager
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	WarnLogger  *log.Logger
	RootFolder  string
}

func NewApp() *App {
	rootFolder := os.Getenv("GOOGLE_ROOT_FOLDER_ID")
	infoLogger := log.New(log.Writer(), "[INFO] ", log.LstdFlags|log.Lshortfile)
	errorLogger := log.New(log.Writer(), "[ERROR] ", log.LstdFlags|log.Lshortfile)
	warnLogger := log.New(log.Writer(), "[WARN] ", log.LstdFlags|log.Lshortfile)

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
			app.handlePostbackEvent(event, user, session)
		default:
			app.WarnLogger.Println("\n\tUnknown event type: ", event.Type)
			app.Bot.SendDefaultReply(event.ReplyToken)

		}
	}
}

func (app *App) handleMessageEvent(event *linebot.Event, user *db.UserData, session *db.UserSession) {
	if user.TestNumber == -1 {
		// Convert the message containing users's test number to an integer
		msg := event.Message.(*linebot.TextMessage).Text
		number, err := strconv.Atoi(msg)
		if err != nil {
			app.WarnLogger.Println("\n\tInvalid test number")
			app.Bot.SendReply(event.ReplyToken, "請於輸入測試編號（2碼）後開始使用！")
			return
		}

		// Update the user's test number
		app.Db.UpdateUserTestNumber(user, number)
		app.Bot.SendReply(event.ReplyToken, "測試編號已設定為"+strconv.Itoa(number))
		return
	}

	switch event.Message.(type) {
	case *linebot.TextMessage:
		app.handleTextMessage(event, user)
	case *linebot.VideoMessage:
		if session.UserState == db.UploadingVideo {
			app.resolveUploadVideo(event, user, session)
		} else {
			app.WarnLogger.Println("\n\tUnknown message type: ", event.Message.Type())
			app.Bot.SendDefaultErrorReply(event.ReplyToken)
		}
	default:
		app.WarnLogger.Println("\n\tUnknown message type: ", event.Message.Type())
		app.Bot.SendDefaultReply(event.ReplyToken)
	}
}

func (app *App) handleTextMessage(event *linebot.Event, user *db.UserData) {
	replyToken := event.ReplyToken
	switch event.Message.(*linebot.TextMessage).Text {
	case "專家影片":
		app.resetUserSession(user.Id)
		_, err := app.Bot.PromptHandednessSelection(replyToken)
		if err != nil {
			app.ErrorLogger.Println("\n\tError prompting handedness selection: ", err)
		}
	case "分析影片":
		app.resetUserSession(user.Id)
		_, err := app.Bot.PromptHandednessSelection(replyToken)
		if err != nil {
			app.ErrorLogger.Println("\n\tError prompting handedness selection: ", err)
		}
		app.updateUserState(user.Id, db.UploadingVideo)
	case "課程大綱":
		app.resetUserSession(user.Id)
		res, err := app.Bot.SendSyllabus(replyToken)
		if err != nil {
			app.WarnLogger.Println("\n\tError sending syllabus: ", err)
		}
		app.InfoLogger.Println("\n\tSyllabus sent. Response from line: ", res)
	}
}

func (app *App) handlePostbackEvent(event *linebot.Event, user *db.UserData, session *db.UserSession) {
	app.InfoLogger.Println("\n\tPostback event:", event.Postback.Data)
	replyToken := event.ReplyToken
	tmp := strings.Split(event.Postback.Data, "&")
	var data [2][2]string
	for i, t := range tmp {
		data[i] = [2]string(strings.SplitN(t, "=", 2))
	}

	if len(data) == 0 {
		app.WarnLogger.Println("\n\tEmpty postback data")
	} else if data[0][0] == "video" {
		var video line.VideoInfo
		json.Unmarshal([]byte(data[0][1]), &video)
		app.Bot.SendVideoMessage(replyToken, video)
	} else if data[0][0] == "handedness" {
		app.handleHandednessReply(replyToken, user, data[0][1], session)
	} else if data[1][0] == "date" {
		app.handleDateReply(data[1][1], replyToken, user, session)
	} else {
		app.handleUserAction(event, user, data)
	}
}

func (app *App) handleDateReply(date string, replyToken string, user *db.UserData, session *db.UserSession) {
	app.Db.UpdateUserSession(user.Id, db.UserSession{
		UserState:    session.UserState,
		UpdatingDate: date,
		Skill:        session.Skill,
	})

	msg := "請輸入【" + date + "】的【" + line.SkillStrToEnum(session.Skill).ChnString() + "】的"
	if session.UserState == db.WritingPreviewNote {
		msg += "課前檢視要點"
	} else {
		msg += "學習反思"
	}
	app.Bot.SendReply(replyToken, msg)
}

func (app *App) handleHandednessReply(replyToken string, user *db.UserData, data string, session *db.UserSession) {
	handedness, err := db.HandednessStrToEnum(data)
	if err != nil {
		app.WarnLogger.Println("\n\tInvalid handedness data")
		app.Bot.SendReply(replyToken, "請選擇左手或右手")
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

	// check line action
	if session.UserState == db.UploadingVideo {
		app.Bot.PromptSkillSelection(replyToken, line.AnalyzeVideo, "請選擇要分析的動作")
	} else {
		app.Bot.PromptSkillSelection(replyToken, line.ViewExpertVideo, "請選擇要觀看的動作")
	}
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
	case line.AddReflection, line.AddPreviewNote:
		// update user session
		var userState db.UserState
		if action.Type == line.AddReflection {
			userState = db.WritingReflection
		} else {
			userState = db.WritingPreviewNote
		}
		go app.Db.UpdateUserSession(user.Id, db.UserSession{
			UserState: userState,
			Skill:     action.Skill.String(),
		})

		err := app.Bot.ResolveViewPortfolio(event, user, action.Skill, userState)
		if err != nil {
			return errors.New("\n\tError resolving view portfolio: " + err.Error())
		}
	case line.ViewPortfolio:
		err := app.Bot.ResolveViewPortfolio(event, user, action.Skill, db.None)
		if err != nil {
			return errors.New("\n\tError resolving view portfolio: " + err.Error())
		}
	case line.ViewExpertVideo:
		err := app.Bot.ResolveViewExpertVideo(event, user, action.Skill)
		if err != nil {
			return errors.New("\n\tError resolving view expert video: " + err.Error())
		}
	case line.AnalyzeVideo:
		// update user session
		go app.Db.UpdateUserSession(user.Id, db.UserSession{
			UserState: db.UploadingVideo,
			Skill:     action.Skill.String(),
		})

		err := app.Bot.PromptUploadVideo(event, user, action.Skill)
		if err != nil {
			return errors.New("\n\tError resolving upload: " + err.Error())
		}
	default:
		return errors.New("\n\tInvalid user action type")
	}
	return nil
}
