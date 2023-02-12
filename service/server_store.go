package service

import (
	"chat-system/pb"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

type UserStore interface {
	SaveUser(user *pb.User)
}
type GroupStore interface {
	JoinGroup(group_data string, user_data *pb.User) (*pb.Group, error)
	CreateGroup(group_data *pb.Group)
	AppendMessage(message_details *pb.AppendChat) (*pb.AppendResponse, error)
	LikeMessage(like_data *pb.LikeMessage) error
	UnLikeMessage(unlike_data *pb.LikeMessage) error
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

	fmt.Printf("user saved. New map Instance : ", userstore)
}

func (group_master *InMemoryGroupStore) CreateGroup(group_data *pb.Group) {
	group_master.Group[group_data.Groupname] = group_data
	log.Printf("New group %v is created", group_data.Groupname)
}

func (group_master *InMemoryGroupStore) JoinGroup(group_name string, user_data *pb.User) (*pb.Group, error) {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	err := group_master.Group[group_name]
	log.Println(err)
	if err != nil {
		group_master.Group[group_name].Participants[user_data.Id] = user_data.Name
	} else {
		new_group := &pb.Group{
			GroupID:      uuid.New().String(),
			Groupname:    group_name,
			Participants: make(map[string]string),
			Messages:     make(map[string]*pb.MessageDetails),
		}
		group_master.CreateGroup(new_group)
		log.Println(group_master)
		group_master.Group[group_name].Participants[user_data.Id] = user_data.Name
	}
	log.Println("In JoinGroup Serve")
	log.Println(group_master.Group[group_name].Participants[user_data.Id])
	return group_master.Group[group_name], nil

}

func (group_master *InMemoryGroupStore) AppendMessage(message_details *pb.AppendChat) (*pb.AppendResponse, error) {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	Message_id := strconv.Itoa(len(group_master.Group[message_details.Group.Groupname].Messages))
	messagedata := &pb.ChatMessage{
		Text:  message_details.Message.Text,
		Likes: make(map[string]string),
	}
	message_data := &pb.MessageDetails{
		User:        message_details.User,
		Messagedata: messagedata,
	}
	group_master.Group[message_details.Group.Groupname].Messages[Message_id] = message_data
	response := &pb.AppendResponse{
		Id: Message_id,
	}
	log.Println(group_master.Group[message_details.Group.Groupname].Messages)
	return response, nil
}

func (group_master *InMemoryGroupStore) LikeMessage(like_data *pb.LikeMessage) error {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	if group_master.Group[like_data.Group.Groupname].Messages[like_data.Messageid] != nil {
		if like_data.User.Name == group_master.Group[like_data.Group.Groupname].Messages[like_data.Messageid].User.Name {
			return errors.New("Cannot like your own message")
		} else {
			if group_master.Group[like_data.Group.Groupname].Messages[like_data.Messageid].Messagedata.Likes[like_data.User.Id] == "" {
				group_master.Group[like_data.Group.Groupname].Messages[like_data.Messageid].Messagedata.Likes[like_data.User.Id] = like_data.User.Name
				log.Println(len(group_master.Group[like_data.Group.Groupname].Messages[like_data.Messageid].Messagedata.Likes))
				return nil
			} else {
				return errors.New("Message already liked.")
			}
		}
	} else {
		return errors.New("Please provide valid message id")
	}

}
func (group_master *InMemoryGroupStore) UnLikeMessage(unlike_data *pb.LikeMessage) error {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	if group_master.Group[unlike_data.Group.Groupname].Messages[unlike_data.Messageid] != nil {
		if unlike_data.User.Name == group_master.Group[unlike_data.Group.Groupname].Messages[unlike_data.Messageid].User.Name {
			return errors.New("Cannot unlike your own message")
		} else {
			if group_master.Group[unlike_data.Group.Groupname].Messages[unlike_data.Messageid].Messagedata.Likes[unlike_data.User.Id] != "" {
				delete(group_master.Group[unlike_data.Group.Groupname].Messages[unlike_data.Messageid].Messagedata.Likes, unlike_data.User.Id)
				log.Println(len(group_master.Group[unlike_data.Group.Groupname].Messages[unlike_data.Messageid].Messagedata.Likes))
				return nil
			} else {
				return errors.New("Cannot unlike a message which isn't liked initially")
			}
		}
	} else {
		return errors.New("Please provide valid message id.")
	}

}
