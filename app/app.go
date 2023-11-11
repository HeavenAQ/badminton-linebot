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
		errorLogger.Println("Error initializing firebase database client:", err)
	}

	bot, err := line.NewLineBotHandler()
	if err != nil {
		errorLogger.Println("Error initializing line bot client:", err)
	}

	drive, err := drive.NewGoogleDriveHandler()
	if err != nil {
		errorLogger.Println("Error initializing google drive client:", err)
	}

	infoLogger.Println("App initialized successfully.")
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
		app.ErrorLogger.Println("Error retrieving callback event:", err)
		return
	}

	// handle events
	log.Println("Handling events...")
	for _, event := range events {
		// get user
		user := app.createUserIfNotExist(event.Source.UserID)
		app.InfoLogger.Println("User: ", user.Name, "sent event: ", event)

		// handler event
		switch event.Type {
		case linebot.EventTypeMessage:
			app.handleMessageEvent(event, user)
		case linebot.EventTypePostback:
			app.handlePostbackEvent(event, user)
		default:
			app.WarnLogger.Println("Unknown event type: ", event.Type)
			app.Bot.SendDefaultReply(event.ReplyToken)

		}
	}
}

func (app *App) handleMessageEvent(event *linebot.Event, user *db.UserData) {
	replyToken := event.ReplyToken
	switch event.Message.(*linebot.TextMessage).Text {
	case "使用說明": // A
		app.Bot.SendInstruction(replyToken)
	case "我的學習歷程": // B
		app.Bot.PromptSkillSelection(replyToken, line.ViewPortfolio, "請選擇要查看的學習歷程")
	case "專家影片": // C
		app.Bot.PromptHandednessSelection(replyToken)
	case "上傳錄影": // D
	// handleCourseInfo(event)
	case "新增學習反思": // E
		app.Bot.PromptSkillSelection(replyToken, line.AddReflection, "請選擇要新增學習反思的動作")
	case "課程大綱": // F
		app.Bot.SendSyllabus(replyToken)
	default:

	}
}

func (app *App) handlePostbackEvent(event *linebot.Event, user *db.UserData) {
	replyToken := event.ReplyToken
	tmp := strings.Split(event.Postback.Data, "&")
	var data [2][2]string
	for i, t := range tmp {
		data[i] = [2]string(strings.SplitN(t, "=", 2))
	}

	if len(data) == 0 {
		app.WarnLogger.Println("Empty postback data")
	} else if data[0][0] == "handedness" {
		app.handleHandednessReply(replyToken, user, data[0][1])
	} else {
		app.handleUserAction(event, user, data)
	}
}

func (app *App) handleHandednessReply(replyToken string, user *db.UserData, data string) {
	handedness, err := db.HandednessStrToEnum(data)
	if err != nil {
		app.WarnLogger.Println("Invalid handedness data")
		app.Bot.SendWrongHandednessReply(replyToken)
		return
	}

	if user.Handedness != handedness {
		err = app.Db.UpdateUserHandedness(user, handedness)
		if err != nil {
			app.WarnLogger.Println("Error updating user handedness:", err)
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
		app.WarnLogger.Println("Invalid postback data")
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}
	err = app.ResolveUserAction(event, user, userAction)
	if err != nil {
		app.ErrorLogger.Println("Error resolving user action:", err)
		app.Bot.SendDefaultErrorReply(replyToken)
		return
	}
}

func (app *App) ResolveUserAction(event *linebot.Event, user *db.UserData, action line.UserActionPostback) error {
	switch action.Type {
	case line.Upload:
		app.Bot.ResolveUpload(event, user, action.Skill)
	case line.AddReflection:
		app.Bot.ResolveAddReflection(event, user, action.Skill)
	case line.ViewPortfolio:
		err := app.Bot.ResolveViewPortfolio(event, user, action.Skill)
		if err != nil {
			return errors.New("Error resolving view portfolio: " + err.Error())
		}
	case line.ViewExpertVideo:
		err := app.Bot.ResolveViewExpertVideo(event, user, action.Skill)
		if err != nil {
			return errors.New("Error resolving view expert video: " + err.Error())
		}
	}
	return errors.New("Invalid user action type")
}
