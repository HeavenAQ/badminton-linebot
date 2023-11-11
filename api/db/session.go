package db

func (handler *FirebaseHandler) GetSession(userId string) (UserData, error) {
	doc, err := handler.GetCollection().Doc(userId).Get(handler.ctx)
	if err != nil {
		return UserData{}, err
	}
	var data UserData
	err = doc.DataTo(&data)
	if err != nil {
		return UserData{}, err
	}
	return data, nil
}
