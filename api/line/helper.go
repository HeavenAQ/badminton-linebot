package line

import (
	"errors"
	"fmt"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) getCarouselItem(work db.Work) *linebot.BubbleContainer {
	return &linebot.BubbleContainer{
		Type: "bubble",
		Header: &linebot.BoxComponent{
			Type:   "box",
			Layout: "vertical",
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   "text",
					Weight: "bold",
					Size:   "xl",
					Text:   work.DateTime,
				},
			},
		},
		Hero: &linebot.ImageComponent{
			Type:        "image",
			URL:         work.Thumbnail,
			Size:        "full",
			AspectRatio: "2:1",
		},
		Body: &linebot.BoxComponent{
			Type:   "box",
			Layout: "horizontal",
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type: "text",
					Size: "md",
					Text: work.Reflection,
					Wrap: true,
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:   "box",
			Layout: "horizontal",
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   "button",
					Style:  "primary",
					Action: linebot.NewURIAction("觀看影片", work.Video),
				},
			},
		},
	}
}

func (handler *LineBotHandler) getCarousels(works map[string]db.Work) ([]*linebot.FlexMessage, error) {
	items := []*linebot.BubbleContainer{}
	carouselItems := []*linebot.FlexMessage{}
	for _, work := range works {
		// if no video or reflection, return error
		if work.Video == "" {
			return nil, errors.New("請上傳" + work.DateTime + "的學習錄影")
		} else if work.Reflection == "" {
			return nil, errors.New("請輸入" + work.DateTime + "的學習反思")
		}

		items = append(items, handler.getCarouselItem(work))

		// since the carousel can only contain 10 items, we need to split the works into multiple carousels in order to display all of them
		if len(items) == 10 {
			carouselItems = append(carouselItems,
				linebot.NewFlexMessage("portfolio",
					&linebot.CarouselContainer{
						Type:     "carousel",
						Contents: items,
					},
				),
			)
			items = []*linebot.BubbleContainer{}
		}
	}
	return carouselItems, nil
}

func (handler *LineBotHandler) replyViewPortfolioError(works map[string]db.Work, event *linebot.Event, skill Skill) error {
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
	return nil
}

func (handler *LineBotHandler) getActionUrls(hand db.Handedness, skill Skill) []string {
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
	return actionUrls[hand][skill]
}

func (handler *LineBotHandler) GetVideoContent(event *linebot.Event) (*linebot.MessageContentResponse, error) {
	msg := event.Message.(*linebot.VideoMessage)
	content, err := handler.bot.GetMessageContent(msg.ID).Do()
	if err != nil {
		return nil, err
	}
	return content, nil
}
