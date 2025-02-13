package db

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
)

type FirebaseHandler struct {
	dbClient *firestore.Client
	ctx      context.Context
}

type UserSession struct {
	Skill        string    `json:"skill"`
	UpdatingDate string    `json:"updatingDate"`
	UserState    UserState `json:"userState"`
}

type UserState int8

const (
	WritingReflection = iota
	WritingPreviewNote
	UploadingVideo
	None
)

type UserData struct {
	Portfolio  Portfolio  `json:"portfolio"`
	FolderIds  FolderIds  `json:"folderIds"`
	Name       string     `json:"name"`
	Id         string     `json:"id"`
	TestNumber int        `json:"testNumber"`
	Handedness Handedness `json:"handedness"`
}

type FolderIds struct {
	Root  string `json:"root"`
	Serve string `json:"serve"`
	Smash string `json:"smash"`
	Clear string `json:"clear"`
}

type Portfolio struct {
	Serve map[string]Work `json:"serve"`
	Smash map[string]Work `json:"smash"`
	Clear map[string]Work `json:"clear"`
}

func (p *Portfolio) GetSkillPortfolio(skill string) map[string]Work {
	switch skill {
	case "serve":
		return p.Serve
	case "smash":
		return p.Smash
	case "clear":
		return p.Clear
	default:
		return nil
	}
}

type Work struct {
	DateTime      string  `json:"date"`
	Thumbnail     string  `json:"thumbnail"`
	SkeletonVideo string  `json:"video"`
	Reflection    string  `json:"reflection"`
	PreviewNote   string  `json:"previewNote"`
	AINote        string  `json:"aiNote"`
	Rating        float32 `json:"rating"`
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

func HandednessStrToEnum(str string) (Handedness, error) {
	switch str {
	case "left":
		return Left, nil
	case "right":
		return Right, nil
	default:
		return -1, errors.New("invalid handedness")
	}
}
