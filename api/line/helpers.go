package line

import (
	"fmt"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) getSkillQuickReplyItems(actionType Action) *linebot.QuickReplyItems {
	items := []*linebot.QuickReplyButton{}
	for _, skill := range []Skill{Lift, Drop, Netplay, Clear, Footwork} {
		items = append(items, linebot.NewQuickReplyButton(
			"",
			linebot.NewPostbackAction(
				skill.String(),
				fmt.Sprintf("type=%s&skill=%s", actionType.String(), skill.String()),
				skill.String(),
				skill.ChnString(),
				"openKeyboard",
				"",
			),
		))
	}
	return linebot.NewQuickReplyItems(items...)
}
