package db

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
)

type UserSessionMap map[string]UserSession
type UserSession struct {
	UserState UserState `json:"userState"`
	Skill     string    `json:"skill"`
}

type UserState int8

const (
	WritingReflection = iota
	UploadingVideo
)

type FirebaseHandler struct {
	dbClient *firestore.Client
	ctx      context.Context
}

type UserMap map[string]UserData
type UserData struct {
	Name       string     `json:"name"`
	Handedness Handedness `json:"handedness"`
	Id         string     `json:"id"`
	FolderIds  FolderIds  `json:"folderIds"`
	Portfolio  Portfolio  `json:"portfolio"`
}

type FolderIds struct {
	Root     string `json:"root"`
	Lift     string `json:"lift"`
	Drop     string `json:"drop"`
	Netplay  string `json:"netplay"`
	Clear    string `json:"clear"`
	Footwork string `json:"footwork"`
}

type Portfolio struct {
	Lift     map[string]Work `json:"lift"`
	Drop     map[string]Work `json:"drop"`
	Netplay  map[string]Work `json:"netplay"`
	Clear    map[string]Work `json:"clear"`
	Footwork map[string]Work `json:"footwork"`
}

func (p *Portfolio) GetSkillMap(skill string) map[string]Work {
	switch skill {
	case "lift":
		return p.Lift
	case "drop":
		return p.Drop
	case "netplay":
		return p.Netplay
	case "clear":
		return p.Clear
	case "footwork":
		return p.Footwork
	default:
		return nil
	}
}

type Work struct {
	Date       string `json:"date"`
	Video      string `json:"video"`
	Thumbnail  string `json:"thumbnail"`
	Reflection string `json:"reflection"`
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
