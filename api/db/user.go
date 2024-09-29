package db

import (
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
			Root:  userFolders.RootFolderId,
			Serve: userFolders.ServeFolderId,
			Smash: userFolders.SmashFolderId,
			Clear: userFolders.ClearFolderId,
		},
		Portfolio: Portfolio{
			Serve: map[string]Work{},
			Smash: map[string]Work{},
			Clear: map[string]Work{},
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

func (handler *FirebaseHandler) CreateUserPortfolioVideo(user *UserData, userPortfolio *map[string]Work, session *UserSession, driveFile *googleDrive.File, thumbnailFile *googleDrive.File, aiRating float32, aiSuggestions string) error {
	id := driveFile.Id
	date := driveFile.Name
	work := Work{
		DateTime:      date,
		Rating:        aiRating,
		Reflection:    "尚未填寫心得",
		PreviewNote:   "尚未填寫課前檢視要點",
		AINote:        aiSuggestions,
		SkeletonVideo: id,
		Thumbnail:     thumbnailFile.Id,
	}
	(*userPortfolio)[date] = work
	handler.UpdateUserSession(user.Id, *session)
	return handler.updateUserData(user)
}

func (handler *FirebaseHandler) UpdateUserPortfolioReflection(user *UserData, userPortfolio *map[string]Work, session *UserSession, reflection string) error {
	targetWork := (*userPortfolio)[session.UpdatingDate]
	work := Work{
		DateTime:      targetWork.DateTime,
		Rating:        targetWork.Rating,
		Reflection:    reflection,
		PreviewNote:   targetWork.PreviewNote,
		SkeletonVideo: targetWork.SkeletonVideo,
		Thumbnail:     targetWork.Thumbnail,
		AINote:        targetWork.AINote,
	}
	(*userPortfolio)[session.UpdatingDate] = work

	err := handler.UpdateUserSession(user.Id, *session)
	if err != nil {
		return err
	}
	return handler.updateUserData(user)
}

func (handler *FirebaseHandler) UpdateUserPortfolioPreviewNote(user *UserData, userPortfolio *map[string]Work, session *UserSession, previewNote string) error {
	targetWork := (*userPortfolio)[session.UpdatingDate]
	work := Work{
		DateTime:      targetWork.DateTime,
		Reflection:    targetWork.Reflection,
		Rating:        targetWork.Rating,
		AINote:        targetWork.AINote,
		PreviewNote:   previewNote,
		SkeletonVideo: targetWork.SkeletonVideo,
		Thumbnail:     targetWork.Thumbnail,
	}
	(*userPortfolio)[session.UpdatingDate] = work

	err := handler.UpdateUserSession(user.Id, *session)
	if err != nil {
		return err
	}
	return handler.updateUserData(user)
}
