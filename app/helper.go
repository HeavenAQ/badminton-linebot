package app

import "github.com/HeavenAQ/api/db"

func (app *App) createUser(userId string) db.UserData {
	username, err := app.Bot.GetUserName(userId)
	if err != nil {
		app.ErrorLogger.Println("Error getting new user's name:", err)
	}
	userFolders, err := app.Drive.CreateUserFolders(userId, username)
	if err != nil {
		app.ErrorLogger.Println("Error creating new user's folders:", err)
	}
	userData, err := app.Db.CreateUserData(userFolders)
	if err != nil {
		app.ErrorLogger.Println("Error creating new user's data:", err)
	}
	return userData[userId]
}

func (app *App) createUserIfNotExist(userId string) (user db.UserData) {
	userData, err := app.Db.GetUserData(userId)
	if err != nil {
		app.WarnLogger.Println("User not found, creating new user...")
		user = app.createUser(userId)
	} else {
		user = userData[userId]
	}
	return
}
