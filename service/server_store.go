package service

import (
	"chat-system/pb"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
)

type UserStore interface {
	SaveUser(user *pb.User)
}
type GroupStore interface {
	GetGroup(groupname string) (*pb.Group, error)
	JoinGroup(groupname string, user *pb.User) (*pb.Group, error)
	AppendMessage(message_details *pb.AppendChat) error
	LikeMessage(like *pb.LikeMessage) error
	UnLikeMessage(unlike *pb.UnLikeMessage) error
	
}

type InMemoryUserStore struct {
	mutex sync.RWMutex
	User  map[string]*pb.User
}

type InMemoryGroupStore struct {
	mutex sync.RWMutex
	Group map[string]*pb.Group
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		User: make(map[string]*pb.User),
	}
}

func NewInMemoryGroupStore() *InMemoryGroupStore {
	return &InMemoryGroupStore{
		Group: make(map[string]*pb.Group),
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

	fmt.Printf("user saved. New map Instance : %s", userstore)
}

func (group_master *InMemoryGroupStore) GetGroup(groupname string) (*pb.Group, error){
	return group_master.Group[groupname] , nil
}

func (group_master *InMemoryGroupStore) JoinGroup(groupname string, user *pb.User) (*pb.Group, error) {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	_, found := group_master.Group[groupname]
	if found {
		group_master.Group[groupname].Participants[user.Id] = user.Name
		return group_master.Group[groupname], nil
	}
	new_group := &pb.Group{
		GroupID:      uuid.New().String(),
		Groupname:    groupname,
		Participants: make(map[string]string),
		Messages:     make(map[string]string),
		Likes:        make(map[string]int32),
	}
	group_master.Group[groupname] = new_group
	group_master.Group[groupname].Participants[user.Id] = user.Name

	return group_master.Group[groupname], nil
}

func (group_master *InMemoryGroupStore) AppendMessage(appendchat *pb.AppendChat) error  {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	msgId := uuid.New().String()
	groupname := appendchat.Group.GetGroupname()
	message := appendchat.Message.GetText()
	group_master.Group[groupname].Messages[msgId] = message

	log.Println(group_master.Group[groupname].GetMessages())
	return nil
}

func (group_master *InMemoryGroupStore) LikeMessage(likemessage *pb.LikeMessage) error {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	groupname := likemessage.Group.GetGroupname()
	msgId := likemessage.GetMessageid()
	group_master.Group[groupname].Likes[msgId]++

	log.Println(group_master.Group[groupname].GetLikes())
	return nil
}
func (group_master *InMemoryGroupStore) UnLikeMessage(unlikemessage *pb.UnLikeMessage) error {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	groupname := unlikemessage.Group.GetGroupname()
	msgId := unlikemessage.GetMessageid()
	if group_master.Group[groupname].Likes[msgId] > 0 {
		group_master.Group[groupname].Likes[msgId]--
	}

	log.Println(group_master.Group[groupname].GetLikes())
	return nil
}
