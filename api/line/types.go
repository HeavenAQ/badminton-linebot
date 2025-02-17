package line

import (
	"errors"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type LineBotHandler struct {
	bot *linebot.Client
}

type enum interface {
	String() string
	ChnString() string
}

type Skill int8

const (
	Serve Skill = iota
	Smash
	Clear
)

func (s Skill) String() string {
	return [...]string{"serve", "smash", "clear"}[s]
}

func (s Skill) ChnString() string {
	return [...]string{"發球", "殺球", "高遠球"}[s]
}

func SkillStrToEnum(str string) Skill {
	switch str {
	case "serve":
		return Serve
	case "smash":
		return Smash
	case "clear":
		return Clear
	default:
		return -1
	}
}

type Action int8

const (
	AnalyzeVideo Action = iota
	AddReflection
	ViewPortfolio
	AddPreviewNote
	ViewInstruction
	ViewExpertVideo
)

func (a Action) String() string {
	return [...]string{"analyze_video", "add_reflection", "view_portfolio", "add_preview_note", "view_instruction", "view_expert_video"}[a]
}

func (a Action) ChnString() string {
	return [...]string{"上傳影片", "本週學習反思", "學習歷程", "課前動作檢測", "使用說明", "專家影片"}[a]
}

func ActionStrToEnum(str string) Action {
	switch str {
	case "analyze_video":
		return AnalyzeVideo
	case "add_reflection":
		return AddReflection
	case "view_portfolio":
		return ViewPortfolio
	case "view_expert_video":
		return ViewExpertVideo
	case "add_preview_note":
		return AddPreviewNote
	case "view_instruction":
		return ViewInstruction
	default:
		return -1
	}
}

type UserActionPostback struct {
	Type  Action `json:"type"`
	Skill Skill  `json:"skill"`
}

func (user *UserActionPostback) String() string {
	actionType := "type=" + user.Type.String()
	skillType := "skill=" + user.Skill.String()
	return actionType + "&" + skillType
}

func (user *UserActionPostback) FromArray(arr [2][2]string) error {
	user.Type = ActionStrToEnum(arr[0][1])
	user.Skill = SkillStrToEnum(arr[1][1])
	if user.Type == -1 || user.Skill == -1 {
		return errors.New("Invalid postback data")
	}
	return nil
}

type CarouselBtn int8

const (
	VideoLink CarouselBtn = iota
	VideoDate
)

type VideoInfo struct {
	VideoId     string `json:"video_id"`
	ThumbnailId string `json:"thumbnail_id"`
}
