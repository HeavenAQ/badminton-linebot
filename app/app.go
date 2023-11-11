package app

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/HeavenAQ/api/db"
	"github.com/HeavenAQ/api/drive"
	"github.com/HeavenAQ/api/line"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type App struct {
	Bot         *line.LineBotHandler
	Drive       *drive.GoogleDriveHandler
	Db          *db.FirebaseHandler
	RootFolder  string
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

func (app *App) handleMessageEvent(event *linebot.Event, user db.UserTemplate) {
	replyToken := event.ReplyToken
	switch event.Message.(*linebot.TextMessage).Text {
	case "使用說明": // A
		app.Bot.SendInstruction(replyToken)
	case "我的學習歷程": // B
	// handleCourseInfo(event)
	case "專家影片": // C
	// handleCourseInfo(event)
	case "上傳錄影": // D
	// handleCourseInfo(event)
	case "新增學習反思": // E
		app.Bot.PromptSelectReflection(replyToken, user)
	case "課程大綱": // F
		app.Bot.SendSyllabus(replyToken)
	default:

	}
}

func (app *App) handlePostbackEvent(event *linebot.Event, user db.UserTemplate) {
	replyToken := event.ReplyToken
	tmp := strings.Split(event.Postback.Data, "&")
	var data []string
	for i, t := range tmp {
		data[i] = strings.Split(t, "=")[1]
	}

}
