package line

import (
	"fmt"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler *LineBotHandler) gePortfolioRating(work db.Work) *linebot.BoxComponent {
	rating := work.Rating
	contents := []linebot.FlexComponent{}
	for i := 0; i < 5; i++ {
		rating -= 20

		var url string
		if rating >= 0 {
			url = "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
		} else {
			url = "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gray_star_28.png"
		}
		contents = append(contents, &linebot.IconComponent{
			Type: "icon",
			Size: "sm",
			URL:  url,
		})
	}
	contents = append(contents, &linebot.TextComponent{
		Type:   "text",
		Text:   fmt.Sprintf("%.1f", work.Rating),
		Size:   "sm",
		Color:  "#8c8c8c",
		Margin: "md",
		Flex:   linebot.IntPtr(0),
	})
	return &linebot.BoxComponent{
		Type:     "box",
		Layout:   "baseline",
		Margin:   "md",
		Contents: contents,
	}
}

func (handler *LineBotHandler) getCarouselItem(work db.Work, userState db.UserState) *linebot.BubbleContainer {
	rating := handler.gePortfolioRating(work)
	var btnAction linebot.TemplateAction
	if userState == db.WritingPreviewNote {
		btnAction = linebot.NewPostbackAction("新增課前動作檢測要點", "type=add_preview_note&date="+work.DateTime, "", "", "openKeyboard", "")
	} else if userState == db.WritingReflection {
		btnAction = linebot.NewPostbackAction("新增學習反思", "type=add_reflection&date="+work.DateTime, "", "", "openKeyboard", "")
	}

	footerContents := []linebot.FlexComponent{
		&linebot.ButtonComponent{
			Type:   "button",
			Style:  "link",
			Height: "sm",
			Action: linebot.NewPostbackAction(
				"查看影片",
				"video_id="+work.SkeletonVideo,
				"",
				"",
				"",
				"",
			),
		},
	}

	if userState != db.None {
		footerContents = append(footerContents, &linebot.ButtonComponent{
			Type:   "button",
			Style:  "link",
			Height: "sm",
			Action: btnAction,
		})
	}

	return &linebot.BubbleContainer{
		Type: "bubble",
		Hero: &linebot.ImageComponent{
			Type:        "image",
			URL:         "https://drive.google.com/thumbnail?authuser=0&sz=w1080&id=" + work.SkeletonVideo,
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
				rating,
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
									Text:   "需調整細節：",
									Color:  "#000000",
									Size:   "md",
									Flex:   linebot.IntPtr(1),
									Weight: "bold",
								},
								&linebot.TextComponent{
									Type:  "text",
									Text:  work.AINote,
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
			Type:     "box",
			Layout:   "vertical",
			Spacing:  "sm",
			Contents: footerContents,
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

func (handler *LineBotHandler) getCarousels(works map[string]db.Work, skill Skill, userState db.UserState) ([]*linebot.FlexMessage, error) {
	items := []*linebot.BubbleContainer{}
	carouselItems := []*linebot.FlexMessage{}
	for _, work := range works {
		items = append(items, handler.getCarouselItem(work, userState))

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
			Serve: []string{
				"https://www.youtube.com/watch?v=lenLFoRFPlk&list=PLZEILcK2CNCvVRym5xnKSFGFHmD13wQhM",
				"https://youtu.be/k9RejtgoatA",
			},
			Smash: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
			Clear: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
		},
		db.Left: {
			Serve: []string{
				"https://youtu.be/ah9ZE9KNFpI",
				"https://youtu.be/JKbQSG27vkk",
			},
			Smash: []string{
				"https://tmp.com",
				"https://tmp.com",
			},
			Clear: []string{
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
