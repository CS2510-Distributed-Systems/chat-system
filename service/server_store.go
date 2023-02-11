package service

import (
	"chat-system/pb"
	"fmt"
	"sync"
)

type UserStore interface {
	SaveUser(user *pb.User) 
	
}
type InMemoryUserStore struct {
	mutex sync.RWMutex
	User  map[string]*pb.User
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		User: make(map[string]*pb.User),
	}
}

func (userstore *InMemoryUserStore) SaveUser(user *pb.User) {
	//find if the user is already present
	//TODO

	//else save the new user
	userstore.mutex.Lock()
	defer userstore.mutex.Unlock()
	Id := user.GetId()
	userstore.User[Id] = user

	fmt.Printf("user saved. New map Instance : ", userstore)
}
