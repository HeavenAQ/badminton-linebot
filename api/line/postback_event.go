package line

import (
	"errors"
	"fmt"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) getSkillQuickReplyItems(actionType Action) *linebot.QuickReplyItems {
	switch actionType {
	case ViewPortfolio:
		return handler.getAddReflectionQuickReplyItems(actionType)
	case ViewExpertVideo:
		return handler.getAddReflectionQuickReplyItems(actionType)
	case Upload:
		return handler.getUploadQuickReplyItems(actionType)
	case AddReflection:
		return handler.getAddReflectionQuickReplyItems(actionType)
	default:
		return nil
	}
}

func (handler *LineBotHandler) getUploadQuickReplyItems(actionType Action) *linebot.QuickReplyItems {
	items := []*linebot.QuickReplyButton{}
	userAction := UserActionPostback{Type: actionType}
	for _, skill := range []Skill{Lift, Drop, Netplay, Clear, Footwork} {
		userAction.Skill = skill
		items = append(items, linebot.NewQuickReplyButton(
			"",
			linebot.NewPostbackAction(
				skill.String(),
				userAction.String(),
				"",
				skill.ChnString(),
				"openCamera",
				"",
			),
		))
	}
	return linebot.NewQuickReplyItems(items...)
}

func (handler *LineBotHandler) getAddReflectionQuickReplyItems(actionType Action) *linebot.QuickReplyItems {
	items := []*linebot.QuickReplyButton{}
	userAction := UserActionPostback{Type: actionType}
	for _, skill := range []Skill{Lift, Drop, Netplay, Clear, Footwork} {
		userAction.Skill = skill
		items = append(items, linebot.NewQuickReplyButton(
			"",
			linebot.NewPostbackAction(
				skill.String(),
				userAction.String(),
				"",
				skill.ChnString(),
				"openKeyboard",
				"",
			),
		))
	}
	return linebot.NewQuickReplyItems(items...)
}

func (handler *LineBotHandler) getHandednessQuickReplyItems() *linebot.QuickReplyItems {
	items := []*linebot.QuickReplyButton{}
	for _, handedness := range []db.Handedness{db.Left, db.Right} {
		linebot.NewQuickReplyButton(
			"",
			linebot.NewPostbackAction(
				handedness.String(),
				handedness.String(),
				"",
				handedness.ChnString(),
				"",
				"",
			),
		)
	}
	return linebot.NewQuickReplyItems(items...)
}

func (handler *LineBotHandler) ResolveViewExpertVideo(event *linebot.Event, user *db.UserData, skill Skill) error {
	actionUrls := map[db.Handedness]map[Skill][]string{
		db.Right: {
			Lift: []string{
				"https://www.youtube.com/watch?v=lenLFoRFPlk&list=PLZEILcK2CNCvVRym5xnKSFGFHmD13wQhM",
				"https://youtu.be/k9RejtgoatA",
			},
			Drop: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
			Netplay: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
			Clear: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
			Footwork: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
		},
		db.Left: {
			Lift: []string{
				"https://youtu.be/ah9ZE9KNFpI",
				"https://youtu.be/JKbQSG27vkk",
			},
			Drop: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
			Netplay: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
			Clear: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
			Footwork: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
		},
	}

	urls := actionUrls[user.Handedness][skill]
	if len(urls) == 0 {
		return errors.New("No expert video found")
	}

	msgs := []linebot.SendingMessage{
		linebot.NewTextMessage(
			fmt.Sprintf(
				"以下為【%v】-【%v】示範影片：",
				user.Handedness.ChnString(),
				skill.ChnString(),
			)),
	}

	for i, url := range urls {
		msg := fmt.Sprintf("專家影片%v：\n%v", i+1, url)
		msgs = append(msgs, linebot.NewTextMessage(msg))
	}

	handler.bot.ReplyMessage(event.ReplyToken, msgs...).Do()
	return nil
}
