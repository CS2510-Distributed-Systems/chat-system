package service

import (
	"chat-system/pb"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type UserStore interface {
	SaveUser(user *pb.User) error
}
type GroupStore interface {
	GetGroup(groupname string) (*pb.Group, error)
	Modified(group *pb.Group, groupname string) bool
	JoinGroup(groupname string, user *pb.User) (*pb.Group, error)
	AppendMessage(appendchat *pb.AppendChat) error
	LikeMessage(like *pb.LikeMessage) error
	UnLikeMessage(unlike *pb.UnLikeMessage) error
}

func (group_master *InMemoryGroupStore) Modified(group *pb.Group, groupname string) bool {
	current_group_instance, err := group_master.GetGroup(groupname)
	if err != nil {
		log.Printf("failed to get the group %s", err)
	}
	log.Println("Latest data: ")
	log.Println(current_group_instance.Messages)

	log.Println("Previous data: ")
	log.Println(group.Messages)
	if current_group_instance != group {
		return true
	}
	return false
}

type InMemoryUserStore struct {
	mutex sync.RWMutex
	User  map[uint32]*pb.User
}

type InMemoryGroupStore struct {
	mutex sync.RWMutex
	Group map[string]*pb.Group
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		User: make(map[uint32]*pb.User),
	}
}

func NewInMemoryGroupStore() *InMemoryGroupStore {

	return &InMemoryGroupStore{
		Group: make(map[string]*pb.Group, 0),
	}
}

func (userstore *InMemoryUserStore) SaveUser(user *pb.User) error {
	userstore.mutex.Lock()
	defer userstore.mutex.Unlock()

	usercopy := &pb.User{}
	err := copier.Copy(usercopy, user)
	if err != nil {
		return fmt.Errorf("error while deepcopy user: %w", err)
	}
	Id := usercopy.GetId()
	userstore.User[Id] = usercopy

	log.Printf("user %v logged in the server", user.GetName())

	return nil
}

func (group_master *InMemoryGroupStore) GetGroup(groupname string) (*pb.Group, error) {
	return group_master.Group[groupname], nil
}

func (group_master *InMemoryGroupStore) JoinGroup(groupname string, user *pb.User) (*pb.Group, error) {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	//try finding the group in the groupstore
	group, found := group_master.Group[groupname]
	if found {
		group.Participants[user.GetId()] = user.GetName()
		return group, nil
	}
	//if not found create one

	new_group := &pb.Group{
		GroupID:      uuid.New().ID(),
		Groupname:    groupname,
		Participants: make(map[uint32]string),
		Messages:     make(map[uint32]*pb.ChatMessage),
	}
	// new_group.Messages[0] = &pb.ChatMessage{
	// 	MessagedBy: user,
	// 	Message: "",
	// 	LikedBy: make(map[uint32]string),
	// }
	group_master.Group[groupname] = new_group
	new_group.Participants[user.GetId()] = user.GetName()

	log.Printf("user %v joined %v group", user.GetName(), groupname)
	return new_group, nil
}

func (group_master *InMemoryGroupStore) AppendMessage(appendchat *pb.AppendChat) error {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	//get groupname and message.
	groupname := appendchat.Group.GetGroupname()
	chatmessage := &pb.ChatMessage{
		MessagedBy: appendchat.Chatmessage.MessagedBy,
		Message:    appendchat.Chatmessage.Message,
		LikedBy:    make(map[uint32]string),
	}
	log.Printf("chatmessage arrived is %v", chatmessage)
	//get group and messagenumber
	group, err := group_master.GetGroup(groupname)
	if err != nil {
		return fmt.Errorf("cannot find the group %v", groupname)
	}
	chatmessagenumber := len(group.Messages)
	//append in the group
	group.Messages[uint32(chatmessagenumber)] = chatmessage
	log.Printf("group messages are %v", group.Messages)
	log.Printf("group %v has a new message appended", groupname)
	return nil
}

func (group_master *InMemoryGroupStore) LikeMessage(likemessage *pb.LikeMessage) error {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	groupname := likemessage.Group.GetGroupname()
	likedmsgnumber := likemessage.GetMessageid()
	likeduser := likemessage.User

	//get the group
	group, err := group_master.GetGroup(groupname)
	if err != nil {
		return fmt.Errorf("cannot find the group %v", groupname)
	}
	//validate and get the message to be liked
	message, found := group.Messages[likedmsgnumber]
	if !found {
		return fmt.Errorf("please enter valid message")
	}
	log.Printf("getting the message : %v", message)
	//like it only if he is not the sender of the message
	if message.MessagedBy == likeduser {
		return fmt.Errorf("cannot like you own message")
	}
	//check if the like is already present
	user, found := message.LikedBy[likeduser.GetId()]
	if found {
		return fmt.Errorf("message already liked")
	}
	//like
	message.LikedBy[likeduser.GetId()] = user

	log.Printf("message liked")
	return nil
}

func (group_master *InMemoryGroupStore) UnLikeMessage(unlikemessage *pb.UnLikeMessage) error {
	group_master.mutex.Lock()
	defer group_master.mutex.Unlock()
	groupname := unlikemessage.Group.GetGroupname()
	unlikedmsgnumber := unlikemessage.GetMessageid()
	unlikeduser := unlikemessage.User

	//get the group
	group, err := group_master.GetGroup(groupname)
	if err != nil {
		return fmt.Errorf("cannot find the group %v", groupname)
	}
	//validate and get the message to be liked
	message, found := group.Messages[unlikedmsgnumber]
	if !found {
		return fmt.Errorf("please enter valid message")
	}
	//like it only if he is not the sender of the message
	if message.MessagedBy == unlikeduser {
		return fmt.Errorf("cannot unlike you own message")
	}
	//check if the like is present
	username, found := message.LikedBy[unlikeduser.GetId()]
	if !found {
		return fmt.Errorf("message never liked")
	}
	//unlike
	delete(message.LikedBy, unlikeduser.GetId())

	log.Printf("user %s unliked a message", username)
	return nil
}
