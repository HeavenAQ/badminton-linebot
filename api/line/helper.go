package line

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/exp/maps"
)

func (handler *LineBotHandler) getCarouselItem(work db.Work, btnType CarouselBtn) *linebot.BubbleContainer {
	var btnAction linebot.TemplateAction
	if btnType == VideoLink {
		// video id will be the last part of the thumbnail url
		id := strings.Split(work.Thumbnail, "=")[3]
		btnAction = linebot.NewPostbackAction("觀看影片", "video_id="+id, "", "", "", "")
	} else if btnType == VideoDate {
		btnAction = linebot.NewPostbackAction("更新心得", "type=update_reflection&date="+work.DateTime, "", "", "openKeyboard", "")
	}
	fmt.Print("work.Thumbnail: ", work.Thumbnail)

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
					Text:   "日期：" + work.DateTime[:10],
				},
			},
		},
		Hero: &linebot.ImageComponent{
			Type:        "image",
			URL:         work.Thumbnail,
			Size:        "full",
			AspectRatio: "20:13",
			AspectMode:  "cover",
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

func (handler *LineBotHandler) getCarousels(works map[string]db.Work, skill Skill, carouselBtn CarouselBtn) ([]*linebot.FlexMessage, error) {
	items := []*linebot.BubbleContainer{}
	carouselItems := []*linebot.FlexMessage{}
	sortedWorks := handler.sortWorks(works)

	for _, work := range sortedWorks {
		items = append(items, handler.getCarouselItem(work, carouselBtn))

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
