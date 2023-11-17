package db

import (
	"time"

	drive "github.com/HeavenAQ/api/drive"
	googleDrive "google.golang.org/api/drive/v3"
)

func (handler *FirebaseHandler) CreateUserData(userFolders *drive.UserFolders) (*UserData, error) {
	ref := handler.GetUsersCollection().Doc(userFolders.UserId)
	newUserTemplate := &UserData{
		Name:       userFolders.UserName,
		Id:         userFolders.UserId,
		Handedness: Right,
		FolderIds: FolderIds{
			Root:     userFolders.RootFolderId,
			Lift:     userFolders.LiftFolderId,
			Drop:     userFolders.DropFolderId,
			Netplay:  userFolders.NetplayFolderId,
			Clear:    userFolders.ClearFolderId,
			Footwork: userFolders.FootworkFolderId,
		},
		Portfolio: Portfolio{
			Lift:     map[string]Work{},
			Drop:     map[string]Work{},
			Netplay:  map[string]Work{},
			Clear:    map[string]Work{},
			Footwork: map[string]Work{},
		},
	}

	_, err := ref.Set(handler.ctx, newUserTemplate)
	if err != nil {
		return nil, err
	}
	return newUserTemplate, nil
}

func (handler *FirebaseHandler) GetUserData(userId string) (*UserData, error) {
	docsnap, err := handler.GetUsersCollection().Doc(userId).Get(handler.ctx)
	if err != nil {
		return nil, err
	}
	user := &UserData{}
	docsnap.DataTo(user)
	return user, nil
}

func (handler *FirebaseHandler) updateUserData(user *UserData) error {
	_, err := handler.GetUsersCollection().Doc(user.Id).Set(handler.ctx, *user)
	if err != nil {
		return err
	}
	return nil
}

func (handler *FirebaseHandler) UpdateUserHandedness(user *UserData, handedness Handedness) error {
	user.Handedness = handedness
	return handler.updateUserData(user)
}

func (handler *FirebaseHandler) UpdateUserPortfolioVideo(user *UserData, userPortfolio *map[string]Work, session *UserSession, driveFile *googleDrive.File) error {
	id := driveFile.Id
	date := time.Now().Format("2006-01-02")
	targetWork := (*userPortfolio)[date]
	work := Work{
		DateTime:   driveFile.Name,
		Reflection: targetWork.Reflection,
		Thumbnail:  "https://drive.google.com/thumbnail?id=" + id,
		Video:      "https://drive.google.com/file/d/" + id + "/view?usp=drive_link",
	}
	(*userPortfolio)[date] = work
	handler.UpdateUserSession(user.Id, *session)
	return handler.updateUserData(user)
}

func (handler *FirebaseHandler) UpdateuserPortfolioReflection(user *UserData, userPortfolio *map[string]Work, session *UserSession, reflection string) error {
	date := time.Now().Format("2006-01-02")
	targetWork := (*userPortfolio)[date]
	work := Work{
		DateTime:   targetWork.DateTime,
		Reflection: reflection,
		Video:      targetWork.Video,
		Thumbnail:  targetWork.Thumbnail,
	}
	(*userPortfolio)[date] = work

	err := handler.UpdateUserSession(user.Id, *session)
	if err != nil {
		return err
	}
	return handler.updateUserData(user)
}