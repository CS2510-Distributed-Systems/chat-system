package service

import (
	"chat-system/pb"
	"context"
	"errors"
	"io"
	"log"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ChatServiceServer struct {
	pb.UnimplementedChatServiceServer
	pb.UnimplementedAuthServiceServer
	groupstore GroupStore
	UserStore  UserStore
	clients    ConnStore
}

func NewChatServiceServer(groupstore GroupStore, userstore UserStore, clients ConnStore) *ChatServiceServer {
	return &ChatServiceServer{
		groupstore: groupstore,
		UserStore:  userstore,
		clients:    clients,
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
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("didnt receive the context properly from the client..")
	}

	groupname := md.Get("groupname")[0]
	userid := md.Get("userid")[0]
	client := [2]string{groupname, userid}
	s.clients.RemoveConn(client)

	currentchat := req.GetJoinchat()
	log.Printf("receive a group join request with name: %s", currentchat.Newgroup)
	// remove if user is already in a group
	s.groupstore.RemoveUser(currentchat.User, groupname)
	s.clients.BroadCast(groupname, &pb.GroupChatResponse{Group: s.groupstore.GetGroup(groupname)})
	// join a group
	group_details, err := s.groupstore.JoinGroup(currentchat.Newgroup, currentchat.User)
	if err != nil {
		log.Printf("Failed to join group %v", err)
	}

	log.Printf("Joined group %s", currentchat.Newgroup)
	res := &pb.JoinResponse{
		Group: group_details,
	}

	return res, nil

}

// streaming rpc
func (s *ChatServiceServer) GroupChat(stream pb.ChatService_GroupChatServer) error {
	//fetching the current group of the client from the rpc context.
	_, cancel := context.WithCancel(stream.Context())
	defer cancel()
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		log.Printf("didnt receive the context properly from the client..")
	}

	groupname := md.Get("groupname")[0]
	userid := md.Get("userid")[0]
	client := [2]string{groupname, userid}
	s.clients.AddConn(stream, client)

	errch := make(chan error)
	listener := make(chan string)
	resp := &pb.GroupChatResponse{
		Group:   s.groupstore.GetGroup(groupname),
		Command: "j",
	}
	s.clients.BroadCast(groupname, resp)
	// waitresponse := make(chan error)
	wg.Add(2)

	go receivestream(stream, s, groupname, listener)
	// go SendBroadcast(groupname, resp, listener, s,&waitresponse)
	// err := <-waitresponse
	go func() error {
		defer wg.Done()
		defer log.Println("send server stream ended")
		for {
			command := <-listener
			if command == "q" {
				resp := &pb.GroupChatResponse{
					Group:   s.groupstore.GetGroup(groupname),
					Command: command,
				}
				s.clients.BroadCast(groupname, resp)
				err := errors.New("Graceful shutdown requested")
				stream.SendMsg(err)
				// waitresponse <- err
				return err
			}

			resp := &pb.GroupChatResponse{
				Group:   s.groupstore.GetGroup(groupname),
				Command: command,
			}
			s.clients.BroadCast(groupname, resp)

		}
	}()

	wg.Wait()
	log.Printf("Stream ended for %v. Please join other group", groupname)
	// go refreshgroup(stream, s, resp)

	return <-errch
}

// func SendBroadcast(groupname string, resp *pb.GroupChatResponse, listener chan string, s *ChatServiceServer,waitresponse *chan error) error {

// }

// var wg *sync.WaitGroup
var mutex = &sync.RWMutex{}

// func refreshgroup(stream pb.ChatService_GroupChatServer, s *ChatServiceServer, resp *pb.GroupChatResponse) {
// 	//Update_Chat_Sent := false

// 	var group_instance = pb.Group{
// 		GroupID:      0,
// 		Groupname:    "",
// 		Participants: make(map[uint32]string),
// 		Messages:     make(map[uint32]*pb.ChatMessage),
// 	}

// 	for {
// 		mutex.Lock()
// 		response := s.groupstore.Modified(group_instance, resp.GetGroup().Groupname)
// 		mutex.Unlock()
// 		if response {
// 			log.Println("In Modified block")
// 			var current_group_instance = s.groupstore.GetGroup(resp.GetGroup().Groupname)
// 			updateCurrentMapInstance(&group_instance, current_group_instance)
// 			resp := &pb.GroupChatResponse{
// 				Group:   current_group_instance,
// 				Command: "refreshed",
// 			}
// 			sendstream(stream, resp)
// 			log.Printf("Refresh group sent the updated group to clients")
// 		}
// 	}
// }

// func updateCurrentMapInstance(group_instance *pb.Group, latest_group_data *pb.Group) {
// 	mutex.Lock()
// 	for key, value := range latest_group_data.Participants {
// 		group_instance.Participants[key] = value
// 	}
// 	mutex.Unlock()
// 	mutex.Lock()
// 	for key, value := range latest_group_data.Messages {
// 		group_instance.Messages[key] = value
// 	}
// 	mutex.Unlock()
// }

func receivestream(stream pb.ChatService_GroupChatServer, s *ChatServiceServer, groupname string, listener chan string) error {
	defer wg.Done()
	defer log.Println("Receive server  stream ended")
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}
		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			log.Printf("error in receiving")
			return logError(status.Errorf(codes.Unknown, "cannot receive stream request: %v", err))
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
			listener <- command

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

			listener <- command

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
			listener <- command

		case *pb.GroupChatRequest_Print:
			command := "p"
			listener <- command

		case *pb.GroupChatRequest_Logout:
			command := "q"
			user := req.GetLogout().User
			s.UserStore.DeleteUser(user)
			s.groupstore.RemoveUser(user, groupname)
			client := [2]string{groupname, strconv.Itoa(int(user.Id))}
			listener <- command
			s.clients.RemoveConn(client)
			log.Printf("user : %v left the chat", user.Name)
			return nil

		default:
			log.Printf("do nothing")
		}

	}
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
