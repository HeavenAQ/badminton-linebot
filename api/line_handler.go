package api

import (
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type LineBotHandler struct {
	bot *linebot.Client
}

func NewLineBotHandler() *LineBotHandler {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
	return &LineBotHandler{
		bot,
	}
}

func (handler *LineBotHandler) RetrieveCbEvent(w http.ResponseWriter, req *http.Request) ([]*linebot.Event, error) {
	cb, err := handler.bot.ParseRequest(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		return nil, err
	}
	return cb, nil
}

func (handler *LineBotHandler) handleEvents(cb []*linebot.Event) {
	for _, event := range cb {
		if event.Type != linebot.EventTypeMessage {
			return
		}

	}
}

func (handler *LineBotHandler) handleTextMessage(event *linebot.Event) {
	switch event.Message.(*linebot.TextMessage).Text {
	case "使用說明": // A
	// handleCourseMenu(event)
	case "我的學習歷程": // B
	// handleCourseInfo(event)
	case "專家影片": // C
	// handleCourseInfo(event)
	case "上傳錄影": // D
	// handleCourseInfo(event)
	case "新增學習反思": // E
	// handleCourseInfo(event)
	case "課程大綱": // F
	// handleCourseInfo(event)
	default:

	}
}

func (handler *LineBotHandler) handlerVideoMessage(event *linebot.Event) {

}
