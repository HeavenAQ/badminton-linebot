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
	UserState    UserState `json:"userState"`
	Skill        string    `json:"skill"`
	UpdatingDate string    `json:"updatingDate"`
}

type UserState int8

const (
	WritingReflection = iota
	WritingPreviewNote
	UploadingVideo
	None
)

type UserData struct {
	Name       string     `json:"name"`
	Handedness Handedness `json:"handedness"`
	Id         string     `json:"id"`
	FolderIds  FolderIds  `json:"folderIds"`
	Portfolio  Portfolio  `json:"portfolio"`
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
	SkeletonVideo string  `json:"video"`
	Rating        float32 `json:"rating"`
	Reflection    string  `json:"reflection"`
	PreviewNote   string  `json:"previewNote"`
	AINote        string  `json:"aiNote"`
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
