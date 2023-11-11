package line

import (
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
		handler.bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("請輸入正確的羽球動作")).Do()
		return nil
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

	_, err := handler.bot.ReplyMessage(event.ReplyToken, msgs...).Do()
	if err != nil {
		return err
	}

	return nil
}

func (handler *LineBotHandler) ResolveViewPortfolio(event *linebot.Event, user *db.UserData, skill Skill) error {
	works := user.Portfolio.GetSkillMap(skill.String())
	if works == nil {
		msg := "請輸入正確的羽球動作"
		handler.bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do()
		return nil
	}

	if len(works) == 0 {
		msg := fmt.Sprintf("尚未上傳【%v】的學習反思及影片", skill.ChnString())
		handler.bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do()
		return nil
	}

	carousels, err := handler.getCarousels(works)
	if err != nil {
		return err
	}

	// turn carousels into sending messages
	var sendMsgs []linebot.SendingMessage
	for _, msg := range carousels {
		sendMsgs = append(sendMsgs, msg)
	}

	_, err = handler.bot.ReplyMessage(event.ReplyToken, sendMsgs...).Do()
	if err != nil {
		return err
	}
	return nil
}

func ResolveUpload(event *linebot.Event, user *db.UserData, skill Skill) error {
	return nil
}

func ResolveAddReflection(event *linebot.Event, user *db.UserData, skill Skill) error {
	return nil
}
