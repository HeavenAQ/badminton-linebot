package line

import (
	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) getCarouselItem(work db.Work, btnType CarouselBtn) *linebot.BubbleContainer {
	var btnAction linebot.TemplateAction
	if btnType == VideoLink {
		btnAction = linebot.NewURIAction("觀看影片", work.RawVideo)
	} else if btnType == VideoDate {
		btnAction = linebot.NewPostbackAction("更新心得", "type=update_reflection&date="+work.DateTime, "", "", "openKeyboard", "")
	}

	return &linebot.BubbleContainer{
		Type: "bubble",
		Hero: &linebot.ImageComponent{
			Type:        "image",
			URL:         work.Thumbnail,
			Size:        "full",
			AspectRatio: "20:13",
			AspectMode:  "cover",
		},
		Body: &linebot.BoxComponent{
			Type:   "box",
			Layout: "vertical",
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type:   "text",
					Text:   "🗓️ " + work.DateTime[:10],
					Weight: "bold",
					Size:   "xl",
				},
				&linebot.BoxComponent{
					Type:    "box",
					Layout:  "vertical",
					Margin:  "lg",
					Spacing: "sm",
					Contents: []linebot.FlexComponent{
						&linebot.BoxComponent{
							Type:    "box",
							Layout:  "vertical",
							Spacing: "sm",
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type:   "text",
									Text:   "課前動作檢測要點：",
									Color:  "#000000",
									Size:   "md",
									Flex:   linebot.IntPtr(1),
									Weight: "bold",
								},
								&linebot.TextComponent{
									Type:  "text",
									Text:  work.PreviewNote,
									Wrap:  true,
									Color: "#666666",
									Size:  "sm",
									Flex:  linebot.IntPtr(5),
								},
							},
						},
					},
				},
				&linebot.BoxComponent{
					Type:    "box",
					Layout:  "vertical",
					Margin:  "lg",
					Spacing: "sm",
					Contents: []linebot.FlexComponent{
						&linebot.BoxComponent{
							Type:    "box",
							Layout:  "vertical",
							Spacing: "sm",
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type:   "text",
									Text:   "學習反思：",
									Color:  "#000000",
									Size:   "md",
									Flex:   linebot.IntPtr(1),
									Weight: "bold",
								},
								&linebot.TextComponent{
									Type:  "text",
									Text:  work.Reflection,
									Wrap:  true,
									Color: "#666666",
									Size:  "sm",
									Flex:  linebot.IntPtr(5),
								},
							},
						},
					},
				},
			},
		},
		Footer: &linebot.BoxComponent{
			Type:    "box",
			Layout:  "vertical",
			Spacing: "sm",
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   "button",
					Style:  "link",
					Height: "sm",
					Action: btnAction,
				},
			},
		},
	}
}

func (handler *LineBotHandler) insertCarousel(carouselItems []*linebot.FlexMessage, items []*linebot.BubbleContainer) []*linebot.FlexMessage {
	return append(carouselItems,
		linebot.NewFlexMessage("portfolio",
			&linebot.CarouselContainer{
				Type:     "carousel",
				Contents: items,
			},
		),
	)

}

func (handler *LineBotHandler) getCarousels(works map[string]db.Work, skill Skill, carouselBtn CarouselBtn) ([]*linebot.FlexMessage, error) {
	items := []*linebot.BubbleContainer{}
	carouselItems := []*linebot.FlexMessage{}
	for _, work := range works {
		items = append(items, handler.getCarouselItem(work, carouselBtn))

		// since the carousel can only contain 10 items, we need to split the works into multiple carousels in order to display all of them
		if len(items) == 10 {
			carouselItems = handler.insertCarousel(carouselItems, items)
			items = []*linebot.BubbleContainer{}
		}
	}

	// insert the last carousel
	carouselItems = handler.insertCarousel(carouselItems, items)
	return carouselItems, nil
}

func (handler *LineBotHandler) replyViewPortfolioError(works map[string]db.Work, event *linebot.Event, msg string) error {
	_, err := handler.bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do()
	if err != nil {
		return err
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
