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
	Upload Action = iota
	AddReflection
	ViewPortfolio
	ViewExpertVideo
)

func (a Action) String() string {
	return [...]string{"upload", "add_reflection", "view_portfolio", "view_syllabus", "view_instruction", "view_expert_video"}[a]
}

func (a Action) ChnString() string {
	return [...]string{"上傳錄影", "新增學習反思", "我的學習歷程", "課程大綱", "使用說明", "專家影片"}[a]
}

func ActionStrToEnum(str string) Action {
	switch str {
	case "upload":
		return Upload
	case "add_reflection":
		return AddReflection
	case "view_portfolio":
		return ViewPortfolio
	case "view_expert_video":
		return ViewExpertVideo
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
