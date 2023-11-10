package main

import (
	"log"
	"net/http"
	"os"

	"github.com/HeavenAQ/api"
)

type App struct {
	Bot         *api.LineBotHandler
	Drive       *api.GoogleDriveHandler
	Db          *api.FirebaseHandler
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

	db, err := api.NewFirebaseHandler()
	if err != nil {
		errorLogger.Println("Error initializing firebase database client:", err)
	}

	bot, err := api.NewLineBotHandler()
	if err != nil {
		errorLogger.Println("Error initializing line bot client:", err)
	}

	drive, err := api.NewGoogleDriveHandler()
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
	events, err := app.Bot.RetrieveCbEvent(w, req)
	if err != nil {
		return
	}

	// handle events
	log.Println("Handling events...")
	for _, event := range events {
		app.InfoLogger.Println("Handling event: ", event)
	}
}
