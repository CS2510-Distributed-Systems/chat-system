package service

import (
	"chat-system/pb"
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/grpc/metadata"
)

type ChatServiceServer struct {
	pb.UnimplementedChatServiceServer

	groupstore GroupStore
}

func NewChatServiceServer(groupstore GroupStore) *ChatServiceServer {
	return &ChatServiceServer{
		groupstore: groupstore,
	}
}

// rpc
func (s *ChatServiceServer) JoinGroup(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	group := req.GetJoinchat()
	log.Printf("receive a group join request with name: %s", group.Groupname)
	group_details, err := s.groupstore.JoinGroup(group.Groupname, group.User)
	if err != nil {
		return nil, fmt.Errorf("error while deepcopy user: %w", err)
	}
	//save the group details in the group store
	if err != nil {
		log.Printf("Failed to join group %v", err)
	}
	log.Printf("Joined group %s", group.Groupname)
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
	go receivestream(stream, s.groupstore, groupname)
	go sendstream(stream, s.groupstore, groupname)
	log.Printf("Refresh group triggered")
	go refreshgroup(stream, s, groupname)

	return <-errch
}

var wg sync.WaitGroup
var mutex = &sync.RWMutex{}

func refreshgroup(stream pb.ChatService_GroupChatServer, s *ChatServiceServer, groupname string) {
	//Update_Chat_Sent := false

	var group_instance = pb.Group{
		GroupID:      0,
		Groupname:    "",
		Participants: make(map[uint32]string),
		Messages:     make(map[uint32]*pb.ChatMessage),
	}

	for {
		mutex.Lock()
		response := s.groupstore.Modified(group_instance, groupname)
		mutex.Unlock()
		if response {
			log.Println("In Modified block")
			var current_group_instance, err = s.groupstore.GetGroup(groupname)
			if err != nil {
				log.Printf("failed to get the group %s", err)
			}
			go updateCurrentMapInstance(&group_instance, current_group_instance)
			sendstream(stream, s.groupstore, groupname)
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

func receivestream(stream pb.ChatService_GroupChatServer, groupstore GroupStore, groupname string) {

	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("error in receiving")
			break
		}
		switch req.GetAction().(type) {
		case *pb.GroupChatRequest_Append:
			appendchat := &pb.AppendChat{
				Group:       req.GetAppend().GetGroup(),
				Chatmessage: req.GetAppend().GetChatmessage(),
			}
			err := groupstore.AppendMessage(appendchat)
			if err != nil {
				log.Printf("cannot save the message %s", err)
			}
			sendstream(stream, groupstore, groupname)
		case *pb.GroupChatRequest_Like:
			group := req.GetLike().Group
			msgId := req.GetLike().Messageid
			user := req.GetLike().User
			likemessage := &pb.LikeMessage{
				Group:     group,
				Messageid: msgId,
				User:      user,
			}
			err := groupstore.LikeMessage(likemessage)
			if err != nil {
				log.Printf("%s", err)
			}
			sendstream(stream, groupstore, groupname)

		case *pb.GroupChatRequest_Unlike:
			group := req.GetUnlike().Group
			msgId := req.GetUnlike().Messageid
			user := req.GetUnlike().User
			unlikemessage := &pb.UnLikeMessage{
				Group:     group,
				Messageid: msgId,
				User:      user,
			}
			groupstore.UnLikeMessage(unlikemessage)
			if err != nil {
				log.Printf("some error occured in unliking the message: %s", err)
			}
			sendstream(stream, groupstore, groupname)

		case *pb.GroupChatRequest_Print:
			fmt.Printf("Need to implement this feature")
			sendstream(stream, groupstore, groupname)

		default:
			log.Printf("do nothing")
		}

	}
}

func sendstream(stream pb.ChatService_GroupChatServer, groupstore GroupStore, groupname string) {
	group, err := groupstore.GetGroup(groupname)
	if err != nil {
		log.Printf("error sending group to client %s", err)
	}
	res := &pb.GroupChatResponse{
		Group: group,
	}
	stream.Send(res)
	log.Printf("sending group data of: %s", groupname)
}
