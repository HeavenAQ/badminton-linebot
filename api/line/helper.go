package line

import (
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/exp/maps"
)

func (handler *LineBotHandler) gePortfolioRating(work db.Work) *linebot.BoxComponent {
	rating := work.Rating
	contents := []linebot.FlexComponent{}
	for i := 0; i < 5; i++ {
		var url string
		if rating >= 20 {
			url = "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
		} else {
			url = "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gray_star_28.png"
		}
		contents = append(contents, &linebot.IconComponent{
			Type: "icon",
			Size: "sm",
			URL:  url,
		})

		rating -= 20
	}
	contents = append(contents, &linebot.TextComponent{
		Type:   "text",
		Text:   fmt.Sprintf("%.2f", work.Rating),
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
	switch userState {
	case db.WritingPreviewNote:
		btnAction = linebot.NewPostbackAction("新增課前動作檢測要點", "type=add_preview_note&date="+work.DateTime, "", "", "openKeyboard", "")
	case db.WritingReflection:
		btnAction = linebot.NewPostbackAction("新增學習反思", "type=add_reflection&date="+work.DateTime, "", "", "openKeyboard", "")
	}

	footerContents := []linebot.FlexComponent{
		&linebot.ButtonComponent{
			Type:   "button",
			Style:  "link",
			Height: "sm",
			Action: linebot.NewPostbackAction(
				"查看影片",
				"video={\"video_id\": \""+work.SkeletonVideo+"\", \"thumbnail_id\": \""+work.Thumbnail+"\"}",
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
			URL:         "https://drive.usercontent.google.com/download?id=" + work.Thumbnail,
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

func (handler *LineBotHandler) sortWorks(works map[string]db.Work) []db.Work {
	workValues := maps.Values(works)
	sort.Slice(workValues, func(i, j int) bool {
		dateTimeI, _ := time.Parse("2006-01-02-15-04", workValues[i].DateTime)
		dateTimeJ, _ := time.Parse("2006-01-02-15-04", workValues[j].DateTime)
		return dateTimeI.After(dateTimeJ)
	})

	sortedWorks := []db.Work{}
	for _, workValue := range workValues {
		sortedWorks = append(sortedWorks, workValue)
	}
	return sortedWorks
}

func (handler *LineBotHandler) getCarousels(works map[string]db.Work, userState db.UserState) ([]*linebot.FlexMessage, error) {
	items := []*linebot.BubbleContainer{}
	carouselItems := []*linebot.FlexMessage{}
	sortedWorks := handler.sortWorks(works)
	for _, work := range sortedWorks {
		items = append(items, handler.getCarouselItem(work, userState))

		// since the carousel can only contain 10 items, we need to split the works into multiple carousels in order to display all of them
		if len(items) == 10 {
			carouselItems = handler.insertCarousel(carouselItems, items)
			items = []*linebot.BubbleContainer{}
		}
	}

	// insert the last carousel
	if len(items) > 0 {
		carouselItems = handler.insertCarousel(carouselItems, items)
	}

	// latest work will be displayed last
	slices.Reverse(carouselItems)
	return carouselItems, nil
}

func (handler *LineBotHandler) replyViewPortfolioError(event *linebot.Event, msg string) error {
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
				"https://youtu.be/uE-EHVX1LrA",
			},
			Smash: []string{
				"https://youtu.be/K7EEhEF2vMo",
			},
			Clear: []string{
				"https://youtu.be/K7EEhEF2vMo",
			},
		},
		db.Left: {
			Serve: []string{
				"https://youtu.be/7i0KvbJ4rEE",
				"https://youtu.be/LiQWE6i3bbI",
			},
			Smash: []string{
				"https://youtu.be/yyjC-xXOsdg",
				"https://youtu.be/AzF44kouBBQ",
			},
			Clear: []string{
				"https://youtu.be/yyjC-xXOsdg",
				"https://youtu.be/AzF44kouBBQ",
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
