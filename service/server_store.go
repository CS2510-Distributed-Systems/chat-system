package service

import "chat-system/pb"

//	type SaveUserData interface {
//		SaveUser(user *pb.User) (*User_Data, error)
//	}
type User_Data struct {
	user_auth map[string]*pb.User
}

func newUserStore() *User_Data {
	return &User_Data{
		user_auth: make(map[string]*pb.User),
	}
}

func (new_user *User_Data) SaveUser(user_data *pb.User) (*User_Data, error) {
	new_user.user_auth[user_data.Id] = user_data
	return new_user, nil
}
