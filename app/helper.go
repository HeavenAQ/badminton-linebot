package app

import "github.com/HeavenAQ/api"

func (app *App) createUser(userId string) map[string]api.UserTemplate {
	username, err := app.Bot.GetUserName(userId)
	if err != nil {
		app.ErrorLogger.Println("Error getting new user's name:", err)
	}
	userFolders, err := app.Drive.CreateUserFolders(userId, username)
	if err != nil {
		app.ErrorLogger.Println("Error creating new user's folders:", err)
	}
	user, err := app.Db.CreateUserData(userFolders)
	if err != nil {
		app.ErrorLogger.Println("Error creating new user's data:", err)
	}
	return user
}
