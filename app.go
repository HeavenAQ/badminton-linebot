package main

import (
	"log"
	"net/http"

	"github.com/HeavenAQ/api"
)

type App struct {
	Bot         *api.LineBotHandler
	Drive       *api.GoogleDriveHandler
	RootFolder  *api.GoogleDriveHandler
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	WarnLogger  *log.Logger
}

func NewApp() *App {

	return &App{
		Bot:         api.NewLineBotHandler(),
		Drive:       api.NewGoogleDriveHandler(),
		InfoLogger:  log.New(log.Writer(), "[INFO] ", log.LstdFlags),
		ErrorLogger: log.New(log.Writer(), "[ERROR] ", log.LstdFlags),
		WarnLogger:  log.New(log.Writer(), "[WARN] ", log.LstdFlags),
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
