package service

import (
	"chat-system/pb"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type ChatServiceServer struct {
	pb.UnimplementedChatServiceServer
	pb.UnimplementedAuthServiceServer
	groupstore GroupStore
	UserStore  UserStore
}


func NewChatServiceServer(groupstore GroupStore, userstore UserStore) *ChatServiceServer {
	return &ChatServiceServer{
		groupstore: groupstore,
		UserStore:  userstore,
	}
}

func (s *ChatServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user_name := req.User.Name
	log.Printf("Logging as: %v", user_name)
	newUser := &pb.User{
		Id:   uuid.New().ID(),
		Name: user_name,
	}
	s.UserStore.SaveUser(newUser)
	res := &pb.LoginResponse{
		User: req.GetUser(),
	}

	return res, nil
}

func (s *ChatServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	s.UserStore.DeleteUser(req.User.User)
	resp := &pb.LogoutResponse{
		Status: true,
	}
	log.Println("User deleted")
	return resp, nil
}


// rpc
func (s *ChatServiceServer) JoinGroup(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	currentchat := req.GetJoinchat()
	log.Printf("receive a group join request with name: %s", currentchat.Groupname)
	// remove if user is already in a group	
	s.groupstore.RemoveUser(currentchat.User,currentchat.Groupname)
	// join a group
	group_details, err := s.groupstore.JoinGroup(currentchat.Groupname, currentchat.User)
	
	if err != nil {
		return nil, fmt.Errorf("error while deepcopy user: %w", err)
	}
	//save the group details in the group store
	if err != nil {
		log.Printf("Failed to join group %v", err)
	}
	log.Printf("Joined group %s", currentchat.Groupname)
	res := &pb.JoinResponse{
		Group: group_details,
	}

	return res, nil

}

// streaming rpc
func (s *ChatServiceServer) GroupChat(stream pb.ChatService_GroupChatServer) error {
	//fetching the current group of the client from the rpc context.
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		log.Printf("didnt receive the context properly from the client..")
	}
	groupname := md.Get("groupname")[0]

	errch := make(chan error)
	resp := &pb.GroupChatResponse{
		Group:   s.groupstore.GetGroup(groupname),
		Command: "j",
	}
	go receivestream(stream, s, groupname)
	log.Printf("Refresh group triggered")
	go refreshgroup(stream, s, resp)

	return <-errch
}

var wg sync.WaitGroup
var mutex = &sync.RWMutex{}

func refreshgroup(stream pb.ChatService_GroupChatServer, s *ChatServiceServer, resp *pb.GroupChatResponse) {
	//Update_Chat_Sent := false

	var group_instance = pb.Group{
		GroupID:      0,
		Groupname:    "",
		Participants: make(map[uint32]string),
		Messages:     make(map[uint32]*pb.ChatMessage),
	}

	for {
		mutex.Lock()
		response := s.groupstore.Modified(group_instance, resp.GetGroup().Groupname)
		mutex.Unlock()
		if response {
			log.Println("In Modified block")
			var current_group_instance = s.groupstore.GetGroup(resp.GetGroup().Groupname)
			go updateCurrentMapInstance(&group_instance, current_group_instance)
			resp := &pb.GroupChatResponse{
				Group: current_group_instance,
				Command: "refreshed",
			}
			sendstream(stream, resp)
			log.Printf("Refresh group sent the updated group to clients")
		}
	}
}
func updateCurrentMapInstance(group_instance *pb.Group, latest_group_data *pb.Group) {
	mutex.Lock()
	for key, value := range latest_group_data.Participants {
		group_instance.Participants[key] = value
	}
	mutex.Unlock()
	mutex.Lock()
	for key, value := range latest_group_data.Messages {
		group_instance.Messages[key] = value
	}
	mutex.Unlock()
}

func receivestream(stream pb.ChatService_GroupChatServer, s *ChatServiceServer, groupname string) {

	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("error in receiving")
			break
		}
		switch req.GetAction().(type) {
		case *pb.GroupChatRequest_Append:
			command := "a"
			appendchat := &pb.AppendChat{
				Group:       req.GetAppend().GetGroup(),
				Chatmessage: req.GetAppend().GetChatmessage(),
			}
			err := s.groupstore.AppendMessage(appendchat)
			if err != nil {
				log.Printf("cannot save the message %s", err)
			}
			group := s.groupstore.GetGroup(groupname)
			resp := &pb.GroupChatResponse{
				Group:   group,
				Command: command,
			}
			sendstream(stream, resp)

		case *pb.GroupChatRequest_Like:
			command := "l"
			group := req.GetLike().Group
			msgId := req.GetLike().Messageid
			user := req.GetLike().User
			likemessage := &pb.LikeMessage{
				Group:     group,
				Messageid: msgId,
				User:      user,
			}
			err := s.groupstore.LikeMessage(likemessage)
			if err != nil {
				log.Printf("%s", err)
			}
			resp := &pb.GroupChatResponse{
				Group:   s.groupstore.GetGroup(groupname),
				Command: command,
			}
			sendstream(stream, resp)

		case *pb.GroupChatRequest_Unlike:
			command := "r"
			group := req.GetUnlike().Group
			msgId := req.GetUnlike().Messageid
			user := req.GetUnlike().User
			unlikemessage := &pb.UnLikeMessage{
				Group:     group,
				Messageid: msgId,
				User:      user,
			}
			s.groupstore.UnLikeMessage(unlikemessage)
			if err != nil {
				log.Printf("some error occured in unliking the message: %s", err)
			}
			resp := &pb.GroupChatResponse{
				Group:   s.groupstore.GetGroup(groupname),
				Command: command,
			}
			sendstream(stream, resp)

		case *pb.GroupChatRequest_Print:
			command := "p"
			resp := &pb.GroupChatResponse{
				Group:   s.groupstore.GetGroup(groupname),
				Command: command,
			}
			sendstream(stream, resp)

		case *pb.GroupChatRequest_Logout:
			command := "q"
			user := req.GetLogout().User
			s.UserStore.DeleteUser(user)
			s.groupstore.RemoveUser(user,groupname)
			group := s.groupstore.GetGroup(groupname)
			resp := &pb.GroupChatResponse{
				Group:   group,
				Command: command,
			}
			sendstream(stream, resp)

		default:
			log.Printf("do nothing")
		}

	}
}

func sendstream(stream pb.ChatService_GroupChatServer, resp *pb.GroupChatResponse) {
	stream.Send(resp)
	log.Printf("group data sent to the stream")
}
