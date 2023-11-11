package line

import (
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type LineBotHandler struct {
	bot *linebot.Client
}

type enum interface {
	String() string
	ChnString() string
	StrToEnum(string) enum
}

type Handedness int8

const (
	Left Handedness = iota
	Right
)

func (h Handedness) String() string {
	return [...]string{"left", "right"}[h]
}

func (h Handedness) ChnString() string {
	return [...]string{"左手", "右手"}[h]
}

func (h Handedness) StrToEnum(str string) Handedness {
	switch str {
	case "left":
		return Left
	case "right":
		return Right
	default:
		return -1
	}
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

func (s Skill) StrToEnum(str string) Skill {
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
	ViewSyllabus
	ViewInstruction
	ViewExpertVideo
)

func (a Action) String() string {
	return [...]string{"upload", "add_reflection", "view_portfolio", "view_syllabus", "view_instruction", "view_expert_video"}[a]
}

func (a Action) ChnString() string {
	return [...]string{"上傳錄影", "新增學習反思", "我的學習歷程", "課程大綱", "使用說明", "專家影片"}[a]
}

func (a Action) StrToEnum(str string) Action {
	switch str {
	case "upload":
		return Upload
	case "add_reflection":
		return AddReflection
	case "view_portfolio":
		return ViewPortfolio
	case "view_syllabus":
		return ViewSyllabus
	case "view_instruction":
		return ViewInstruction
	case "view_expert_video":
		return ViewExpertVideo
	default:
		return -1
	}
}
