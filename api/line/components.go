package line

import (
	"strings"
	"time"

	"github.com/HeavenAQ/api/db"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func userProfileComponent(userProfileImg string) *linebot.BoxComponent {
	image := &linebot.ImageComponent{
		Type:       "image",
		URL:        userProfileImg,
		AspectMode: "cover",
		Size:       "full",
	}

	return &linebot.BoxComponent{
		Type:         "box",
		Layout:       "vertical",
		Contents:     []linebot.FlexComponent{image},
		CornerRadius: "100px",
		Width:        "72px",
		Height:       "72px",
	}
}

func reflectionComponent(userName string, reflection string) *linebot.BoxComponent {
	name := &linebot.SpanComponent{
		Type:   "span",
		Text:   userName,
		Weight: "bold",
		Color:  "#000000",
	}

	content := &linebot.SpanComponent{
		Type: "text",
		Text: reflection,
	}

	blank := &linebot.SpanComponent{
		Type: "span",
		Text: "   ",
	}

	thread := &linebot.TextComponent{
		Type:     "text",
		Contents: []*linebot.SpanComponent{name, blank, content},
		Size:     "md",
		Wrap:     true,
	}

	return &linebot.BoxComponent{
		Type:         "box",
		Layout:       "vertical",
		Contents:     []linebot.FlexComponent{thread},
		Spacing:      "xl",
		PaddingStart: "20px",
	}
}

func dateComponent(date string) *linebot.BoxComponent {
	dateTime, _ := time.Parse("2006-01-02-15-04", date)
	formattedDate := dateTime.Format("Mon, Jan 02, 2006, 15:04")
	dateContent := &linebot.TextComponent{
		Type:  "text",
		Text:  formattedDate,
		Wrap:  true,
		Size:  "sm",
		Color: "#aaaaaa",
	}

	return &linebot.BoxComponent{
		Type:    "box",
		Layout:  "baseline",
		Spacing: "md",
		Contents: []linebot.FlexComponent{
			dateContent,
		},
	}
}

func createViewVideoAction(videoLink string, thumbnailLink string) *linebot.PostbackAction {
	// video videoID will be the element after videoID= in the link
	videoID := strings.Split(videoLink, "id=")[1]
	thumbnailID := strings.Split(thumbnailLink, "id=")[1]
	return linebot.NewPostbackAction("觀看影片", "video={\"video_id\":\""+videoID+"\","+"\"thumbnail_id\":\""+thumbnailID+"\"}", "", "", "", "")
}

func createUpdateReflectionAction(date string) *linebot.PostbackAction {
	return linebot.NewPostbackAction("更新心得", "type=update_reflection&date="+date, "", "", "openKeyboard", "")
}

func portfolioCardComponent(work db.Work, userProfileImg string, userName string) *linebot.BubbleContainer {
	viewVideoBtn := createViewVideoAction(work.Video, work.Thumbnail)
	updateReflectionBtn := createUpdateReflectionAction(work.DateTime)

	return &linebot.BubbleContainer{
		Type: "bubble",
		Hero: &linebot.ImageComponent{
			Type:       "image",
			URL:        work.Thumbnail,
			Size:       "full",
			AspectMode: "cover",
		},
		Body: &linebot.BoxComponent{
			Type:          "box",
			Layout:        "vertical",
			Spacing:       "xl",
			PaddingBottom: "20px",
			Contents: []linebot.FlexComponent{
				&linebot.BoxComponent{
					Type:   "box",
					Layout: "horizontal",
					Contents: []linebot.FlexComponent{
						userProfileComponent(userProfileImg),
						reflectionComponent(userName, work.Reflection),
					},
				},
				dateComponent(work.DateTime),
			},
		},
		Footer: &linebot.BoxComponent{
			Type:    "box",
			Layout:  "vertical",
			Spacing: "md",
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type:   "button",
					Style:  "primary",
					Action: updateReflectionBtn,
				},
				&linebot.ButtonComponent{
					Type:   "button",
					Style:  "link",
					Action: viewVideoBtn,
				},
			},
		},
	}
}
