package service

import (
	"chat-system/pb"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type ServerStore interface{

	SaveUser (user *pb.User) (uint32, error)
}

type InMemoryUserStore struct {
	mutex sync.RWMutex
	data []*pb.User
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		data: make([]*pb.User, 0),
	}
}

func (store *InMemoryUserStore) Save(user *pb.User) (uint32, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
 	
	if store.data[user.Id] != nil {
		return user.Id, nil
	}

	//deep copy
	new_user := &pb.User{}
	err := copier.Copy(new_user, user)
	if err != nil {
		fmt.Errorf("cannot perform the save operation: %w", err)
	}

	Id := uuid.New().ID()
	 store.data[Id] = user
	 return Id, nil

}