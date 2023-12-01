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
	Lift Skill = iota
	Drop
	Netplay
	Clear
	Footwork
)

func (s Skill) String() string {
	return [...]string{"lift", "drop", "netplay", "clear", "footwork"}[s]
}

func (s Skill) ChnString() string {
	return [...]string{"挑球", "切球", "小球", "高遠球", "腳步"}[s]
}

func SkillStrToEnum(str string) Skill {
	switch str {
	case "lift":
		return Lift
	case "drop":
		return Drop
	case "netplay":
		return Netplay
	case "clear":
		return Clear
	case "footwork":
		return Footwork
	default:
		return -1
	}
}

type Action int8

const (
	AnalyzeVideo Action = iota
	AddReflection
	ViewPortfolio
	PreviewCourse
	ViewInstruction
	ViewExpertVideo
)

func (a Action) String() string {
	return [...]string{"analyze_video", "add_reflection", "view_portfolio", "preview_course", "view_instruction", "view_expert_video"}[a]
}

func (a Action) ChnString() string {
	return [...]string{"分析影片", "本週學習反思", "學習歷程", "課前動作檢測", "使用說明", "專家影片"}[a]
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
	case "preview_course":
		return PreviewCourse
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
