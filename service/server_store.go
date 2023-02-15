package service

import (
	"chat-system/pb"
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

type UserStore interface {
	SaveUser(user *pb.User) error
	DeleteUser(User *pb.User)
}

type GroupStore interface {
	GetGroup(groupname string) *pb.Group
	Modified(group pb.Group, groupname string) bool
	JoinGroup(groupname string, user *pb.User) (*pb.Group, error)
	AppendMessage(appendchat *pb.AppendChat) error
	LikeMessage(like *pb.LikeMessage) error
	UnLikeMessage(unlike *pb.UnLikeMessage) error
	RemoveUser(user *pb.User, groupname string)
}

func (group_master *InMemoryGroupStore) Modified(group pb.Group, groupname string) bool {
	group_master.mutex.Lock()
	current_group_instance := group_master.GetGroup(groupname)
	group_master.mutex.Unlock()
	res_Participants := reflect.DeepEqual(group.Participants, current_group_instance.Participants)
	res_Messages := reflect.DeepEqual(group.Messages, current_group_instance.Messages)
	if !res_Participants {
		return true
	} else if !res_Messages {
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

func (userstore *InMemoryUserStore) DeleteUser(user *pb.User) {
	delete(userstore.User, user.Id)
}

func (group_master *InMemoryGroupStore) GetGroup(groupname string) *pb.Group {
	return group_master.Group[groupname]
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
	group := group_master.GetGroup(groupname)
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
	group := group_master.GetGroup(groupname)
	//validate and get the message to be liked
	message, found := group.Messages[likedmsgnumber]
	if !found {
		return fmt.Errorf("please enter valid message")
	}
	log.Printf("getting the message : %v", message)
	//like it only if he is not the sender of the message
	if message.MessagedBy.Id == likeduser.Id {
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
	group := group_master.GetGroup(groupname)
	//validate and get the message to be liked
	message, found := group.Messages[unlikedmsgnumber]
	if !found {
		return fmt.Errorf("please enter valid message")
	}
	//like it only if he is not the sender of the message
	if message.MessagedBy.Id == unlikeduser.Id {
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

func (group_master *InMemoryGroupStore) RemoveUser(user *pb.User, groupname string) {
	groupmap := group_master.Group[groupname]
	if groupmap == nil {
		return
	} else if groupmap.Participants == nil {
		return
	} else if groupmap.Participants[user.Id] == "" {

		return
	}

	delete(group_master.Group[groupname].Participants, user.Id)

}

// braodcaster Model
type Broadcaster interface {
	subscribe() <-chan *pb.GroupChatResponse
	unsubscribe(<-chan *pb.GroupChatResponse)
	getsource()chan *pb.GroupChatResponse
}
type broadcaster struct {
	source       chan *pb.GroupChatResponse
	clients      []chan *pb.GroupChatResponse
	addclient    chan chan *pb.GroupChatResponse
	removeclient chan (<-chan *pb.GroupChatResponse)
}

func (bc *broadcaster) getsource() chan *pb.GroupChatResponse {
	return bc.source
}

func (bc *broadcaster) subscribe() <-chan *pb.GroupChatResponse {
	newclient := make(chan *pb.GroupChatResponse)
	bc.addclient <- newclient
	return newclient
}

func (bc *broadcaster) unsubscribe(channel <-chan *pb.GroupChatResponse) {
	bc.removeclient <- channel
}
func NewBroadcaster(ctx context.Context, source chan *pb.GroupChatResponse) *broadcaster {
	broadcastservice := &broadcaster{
		source:       source,
		clients:      make([]chan *pb.GroupChatResponse, 0),
		addclient:    make(chan chan *pb.GroupChatResponse),
		removeclient: make(chan (<-chan *pb.GroupChatResponse)),
	}
	go broadcastservice.serve(ctx)
	return broadcastservice
}

func (bc *broadcaster) serve(ctx context.Context) {
	log.Printf("initializing the Broadcaster")
	defer func() {
		for _, client := range bc.clients {
			if client != nil {
				close(client)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("context is Done")
			return
		case newclient := <-bc.addclient:
			bc.clients = append(bc.clients, newclient)
			log.Printf("new client subscribed : %v", len(bc.clients))
		case removeclient := <-bc.removeclient:
			for i, ch := range bc.clients {
				if ch == removeclient {
					bc.clients[i] = bc.clients[len(bc.clients)-1]
					bc.clients = bc.clients[:len(bc.clients)-1]
					close(ch)
					break
				}
			}
		case val, ok := <-bc.source:
			if !ok {
				log.Println("Something is fishing. source listened but couldnt broadcast")
				return
			}

			for _, client := range bc.clients {
				log.Println("looping through the subscribed clients. ")
				if client != nil {
					select {
					case client <- val:
						log.Println("Broadcasting to: ")
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}
