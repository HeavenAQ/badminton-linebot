package line

import (
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func NewLineBotHandler() (*LineBotHandler, error) {
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		return nil, err
	}
	return &LineBotHandler{
		bot,
	}, nil
}

func (handler *LineBotHandler) RetrieveCbEvent(w http.ResponseWriter, req *http.Request) ([]*linebot.Event, error) {
	cb, err := handler.bot.ParseRequest(req)
	if err != nil {
		w.WriteHeader(400)
		return nil, err
	}
	return cb, nil
}

func (handler *LineBotHandler) GetUserName(userId string) (string, error) {
	profile, err := handler.bot.GetProfile(userId).Do()
	if err != nil {
		return "", err
	}
	return profile.DisplayName, nil
}
