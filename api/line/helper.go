package line

import (
	"slices"
	"sort"
	"time"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"golang.org/x/exp/maps"
)

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

func (handler *LineBotHandler) getCarousels(works map[string]db.Work, userId string, userName string) ([]*linebot.FlexMessage, error) {
	// sort works by date
	sortedWorks := handler.sortWorks(works)

	// get profile image
	profile, _ := handler.bot.GetProfile(userId).Do()
	userProfileImg := profile.PictureURL

	// create portfolio cards to form a carousel
	items := []*linebot.BubbleContainer{}
	carouselItems := []*linebot.FlexMessage{}
	for _, work := range sortedWorks {
		items = append(items, portfolioCardComponent(work, userProfileImg, userName))

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
			Lift: []string{
				"https://youtu.be/6T6zMCKc6Mw",
				"https://youtu.be/k9RejtgoatA",
				"https://youtu.be/4XVJKG6KwlI",
				"https://youtu.be/g58fyhMkRD4",
			},
			Drop: []string{
				"https://youtu.be/ST5citEQZps",
			},
			Netplay: []string{
				"https://youtu.be/mklLfEWPG_U",
			},
			Clear: []string{
				"https://youtu.be/K7EEhEF2vMo",
			},
			Footwork: []string{
				"https://youtu.be/IPl7-mCESfs",
			},
		},
		db.Left: {
			Lift: []string{
				"https://youtu.be/ah9ZE9KNFpI",
				"https://youtu.be/JKbQSG27vkk",
				"https://youtu.be/ah9ZE9KNFpI",
				"https://youtu.be/JKbQSG27vkk",
			},
			Drop: []string{
				"https://youtu.be/zatTzMKNUgY",
				"https://youtu.be/BKpO9u9Ci14",
			},
			Netplay: []string{
				"https://youtu.be/lWnLgTaiSAY",
				"https://youtu.be/KkAfJBuYx00",
			},
			Clear: []string{
				"https://youtu.be/yyjC-xXOsdg",
				"https://youtu.be/AzF44kouBBQ",
			},
			Footwork: []string{
				"https://youtu.be/9i_5PgCYgts",
				"https://youtu.be/AZtvW9faDA8",
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
