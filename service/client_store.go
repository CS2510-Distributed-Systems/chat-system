package service

import (
	"chat-system/pb"

	"google.golang.org/grpc"
)

type ClientStore interface {
	SetUser(*pb.User) error
	SetGroup(*pb.Group) error
	GetUser() *pb.User
	GetGroup() *pb.Group
	GetConn() *grpc.ClientConn
}

type InMemoryClientStore struct {
	active_user *pb.User
	group       *pb.Group
	conn *grpc.ClientConn
}

func NewInMemoryClientStore(conn *grpc.ClientConn) *InMemoryClientStore {
	group := &pb.Group{
		GroupID:      0,
		Groupname:    "",
		Participants: make(map[uint32]string),
		Messages:     make(map[uint32]*pb.ChatMessage),
	}
	active_user := &pb.User{
		Name: "",
		Id:   0,
	}
	
	return &InMemoryClientStore{
		active_user: active_user,
		group:       group,
		conn: conn,
	}
}
func (clientstore *InMemoryClientStore) GetConn() *grpc.ClientConn{
	return clientstore.conn
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
