package line

import (
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) ResolveAddReflection(replyToken string, user db.UserData) {
