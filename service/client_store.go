package service

import (
	"chat-system/pb"
)

type ClientStore interface {
	SetUser(*pb.User) error
	SetGroup(*pb.Group) error
	GetUser() *pb.User
	GetGroup() *pb.Group
}

type InMemoryClientStore struct {
	active_user *pb.User
	group       *pb.Group
}

func NewInMemoryClientStore() *InMemoryClientStore {
	group := &pb.Group{
		GroupID:      "0",
		Groupname:    "None",
		Participants: make(map[string]string),
		Messages:    make(map[string]string),
	}
	return &InMemoryClientStore{
		active_user: &pb.User{
			Name: "None",
			Id:   "0",
		},
		group: group,
	}
}

func (clientstore *InMemoryClientStore) SetUser(user *pb.User) error {
	clientstore.active_user = user
	return nil
}

func (clientstore *InMemoryClientStore) SetGroup(group *pb.Group) error {
	clientstore.group = group
	return nil
}

func (clientstore *InMemoryClientStore) GetUser() *pb.User {
	return clientstore.active_user
}

func (clientstore *InMemoryClientStore) GetGroup() *pb.Group {
	return clientstore.group
}
