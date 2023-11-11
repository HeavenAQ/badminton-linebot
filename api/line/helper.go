package line

import (
	"errors"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func (handler LineBotHandler) getCarouselItem(work db.Work) *linebot.BubbleContainer {
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
					Text:   work.Date,
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

func (handler LineBotHandler) getCarousels(works map[string]db.Work) ([]*linebot.FlexMessage, error) {
	items := []*linebot.BubbleContainer{}
	carouselItems := []*linebot.FlexMessage{}
	for _, work := range works {
		// if no video or reflection, return error
		if work.Video == "" {
			return nil, errors.New("請上傳" + work.Date + "的學習錄影")
		} else if work.Reflection == "" {
			return nil, errors.New("請輸入" + work.Date + "的學習反思")
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
