package line

import (
	"errors"
	"fmt"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) getSkillQuickReplyItems(actionType Action) *linebot.QuickReplyItems {
	items := []*linebot.QuickReplyButton{}
	userAction := UserActionPostback{Type: actionType}
	replyAction := handler.getQuickReplyAction()

	for _, skill := range []Skill{Serve, Smash, Clear} {
		userAction.Skill = skill
		items = append(items, linebot.NewQuickReplyButton(
			"",
			replyAction(userAction),
		))
	}
	return linebot.NewQuickReplyItems(items...)
}

type ReplyAction func(userAction UserActionPostback) linebot.QuickReplyAction

func (handler *LineBotHandler) getQuickReplyAction() ReplyAction {
	return func(userAction UserActionPostback) linebot.QuickReplyAction {
		return linebot.NewPostbackAction(
			userAction.Skill.ChnString(),
			userAction.String(),
			"",
			userAction.Skill.ChnString(),
			linebot.InputOption(""),
			"",
		)
	}
}

func (handler *LineBotHandler) getHandednessQuickReplyItems() *linebot.QuickReplyItems {
	items := []*linebot.QuickReplyButton{}
	for _, handedness := range []db.Handedness{db.Left, db.Right} {
		items = append(items, linebot.NewQuickReplyButton(
			"",
			linebot.NewPostbackAction(
				handedness.ChnString(),
				"handedness="+handedness.String(),
				"",
				handedness.ChnString(),
				"",
				"",
			),
		))
	}
	return linebot.NewQuickReplyItems(items...)
}

func (handler *LineBotHandler) ResolveViewExpertVideo(event *linebot.Event, user *db.UserData, skill Skill) error {
	urlIDs := handler.getActionUrls(user.Handedness, skill)
	if len(urlIDs) == 0 {
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

	for i, url := range urlIDs {
		msg := fmt.Sprintf("專家影片%v：\n%v", i+1, url)
		msgs = append(msgs, linebot.NewTextMessage(msg))
	}

	_, err := handler.bot.ReplyMessage(event.ReplyToken, msgs...).Do()
	if err != nil {
		return err
	}

	return nil
}

func (handler *LineBotHandler) ResolveViewPortfolio(event *linebot.Event, user *db.UserData, skill Skill, userState db.UserState) error {
	// get works from user portfolio
	works := user.Portfolio.GetSkillPortfolio(skill.String())
	if len(works) == 0 {
		var msg string
		if works == nil {
			msg = "請輸入正確的羽球動作"
		} else {
			msg = fmt.Sprintf("尚未上傳【%v】的學習反思及影片", skill.ChnString())
		}

		// reply user with error messages
		handler.replyViewPortfolioError(event, msg)
	}

	// generate carousels from works
	carousels, err := handler.getCarousels(works, userState)
	if err != nil {
		handler.replyViewPortfolioError(event, err.Error())
		return errors.New("\n\tError getting carousels: " + err.Error())
	}

	// turn carousels into sending messages
	var sendMsgs []linebot.SendingMessage
	for _, msg := range carousels {
		sendMsgs = append(sendMsgs, msg)
	}

	_, err = handler.bot.ReplyMessage(event.ReplyToken, sendMsgs...).Do()
	if err != nil {
		handler.replyViewPortfolioError(event, err.Error())
		return err
	}
	return nil
}

func (handler *LineBotHandler) PromptUploadVideo(event *linebot.Event, user *db.UserData, skill Skill) error {
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
