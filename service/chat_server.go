package service

import (
	"chat-system/pb"
	"context"
	"fmt"
	"log"

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

var group_instance = pb.Group{
	GroupID:      0,
	Groupname:    "",
	Participants: make(map[uint32]string),
	Messages:     make(map[uint32]*pb.ChatMessage),
}

// rpc
func (s *ChatServiceServer) JoinGroup(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {

	group := req.GetJoinchat()
	log.Printf("receive a group join request with name: %s", group.Groupname)

	//save the group details in the group store
	group_details, err := s.groupstore.JoinGroup(group.Groupname, group.User)
	group_instance.Groupname = group_details.Groupname
	group_instance.GroupID = group_details.GroupID
	for key, val := range group_details.Participants {
		group_instance.Participants[key] = val
	}
	for key, val := range group_details.Messages {
		group_instance.Messages[key] = val
	}
	//copier.Copy(group_instance, *group_details)
	log.Println(group_instance.Participants)
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
	// group_instance, err := s.groupstore.GetGroup(groupname)
	// if err != nil {
	// 	log.Printf("failed to get the group %s", err)
	// }
	log.Printf("Refresh group triggered")
	go refreshgroup(stream, group_instance, s, groupname)

	return <-errch
}

func refreshgroup(stream pb.ChatService_GroupChatServer, group_instance pb.Group, s *ChatServiceServer, groupname string) {
	//Update_Chat_Sent := false
	//var mutex = &sync.Mutex{}
	for {
		log.Println(group_instance.Participants)
		if s.groupstore.Modified(group_instance, groupname) {
			log.Println("In Modified block")
			current_group_instance, err := s.groupstore.GetGroup(groupname)
			if err != nil {
				log.Printf("failed to get the group %s", err)
			}
			instance := *current_group_instance
			group_instance.Groupname = instance.Groupname
			group_instance.GroupID = instance.GroupID
			for key, val := range instance.Participants {
				group_instance.Participants[key] = val
			}
			for key, val := range instance.Messages {
				group_instance.Messages[key] = val
			}
			//copier.Copy(group_instance, *current_group_instance)
			log.Println(group_instance.Participants)
			sendstream(stream, s.groupstore, groupname)
			log.Printf("Refresh group sent the updated group to clients")
		}
	}
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
