package line

import (
	"fmt"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) getSkillQuickReplyItems(actionType Action) *linebot.QuickReplyItems {
	items := []*linebot.QuickReplyButton{}
	userAction := UserActionPostback{Type: actionType}
	replyAction := handler.getQuickReplyAction(actionType, Lift)

	for _, skill := range []Skill{Lift, Drop, Netplay, Clear, Footwork} {
		userAction.Skill = skill
		items = append(items, linebot.NewQuickReplyButton(
			"",
			replyAction(userAction),
		))
	}
	return linebot.NewQuickReplyItems(items...)
}

type ReplyAction func(userAction UserActionPostback) linebot.QuickReplyAction

func (handler *LineBotHandler) getQuickReplyAction(actionType Action, skill Skill) ReplyAction {
	var inputOption string
	if actionType == AddReflection {
		inputOption = "openKeyboard"
	} else {
		inputOption = ""
	}

	return func(userAction UserActionPostback) linebot.QuickReplyAction {
		return linebot.NewPostbackAction(
			skill.String(),
			userAction.String(),
			"",
			skill.ChnString(),
			linebot.InputOption(inputOption),
			"",
		)
	}
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
	urls := handler.getActionUrls(user.Handedness, skill)
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
	if works == nil || len(works) == 0 {
		handler.replyViewPortfolioError(works, event, skill)
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

func (handler *LineBotHandler) ResolveVideoUpload(event *linebot.Event, user *db.UserData, skill Skill) error {
	_, err := handler.bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage("請上傳影片").WithQuickReplies(
			linebot.NewQuickReplyItems(
				linebot.NewQuickReplyButton(
					"",
					linebot.NewCameraAction("拍攝影片"),
				),
				linebot.NewQuickReplyButton(
					"",
					linebot.NewCameraRollAction("從相簿選擇"),
				),
			),
		),
	).Do()
	return err
}

func (handler *LineBotHandler) ResolveAddReflection(event *linebot.Event, user *db.UserData, skill Skill) error {
	_, err := handler.bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage("請輸入學習反思"),
	).Do()
	return err
}
