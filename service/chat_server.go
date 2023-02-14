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
	groupstore GroupStore
	broadcast  chan *pb.GroupChatResponse
	mutex      sync.RWMutex
	pb.UnimplementedChatServiceServer
}

func NewChatServiceServer(groupstore GroupStore, broadcast chan *pb.GroupChatResponse) *ChatServiceServer {
	return &ChatServiceServer{
		groupstore: groupstore,
		broadcast:  broadcast,
	}
}

var group_instance *pb.Group

// rpc
func (s *ChatServiceServer) JoinGroup(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {

	group := req.GetJoinchat()
	log.Printf("receive a group join request with name: %s", group.Groupname)

	//save the group details in the group store
	group_details, err := s.groupstore.JoinGroup(group.Groupname, group.User)
	if err != nil {
		return nil, fmt.Errorf("error while deepcopy user: %w", err)
	}
	//save the group details in the group store
	log.Println(group_details)
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
	username := md.Get("username")[0]

	errch := make(chan error)

	go s.groupstore.SendBroadcasts(username, stream)
	go receivestream(stream, username, groupname, s)
	log.Printf("Refresh group triggered")
	<-stream.Context().Done()
	return <-errch
}

func refreshgroup(stream pb.ChatService_GroupChatServer, group_instance *pb.Group, s *ChatServiceServer, groupname string) {
	log.Printf("Refresh group started")
	for {
		if s.groupstore.Modified(group_instance, groupname) {
			current_group_instance := s.groupstore.GetGroup(groupname)

			group_instance = current_group_instance
			// sendstream(stream, s.groupstore, groupname)
			log.Printf("Refresh group sent the updated group to clients")
		}
	}
}

func receivestream(stream pb.ChatService_GroupChatServer, username string, groupname string, s *ChatServiceServer) {

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
			err := s.groupstore.AppendMessage(appendchat)
			if err != nil {
				log.Printf("cannot save the message %s", err)
			}

		case *pb.GroupChatRequest_Like:
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
			// sendstream(stream, s.groupstore, groupname)

		case *pb.GroupChatRequest_Unlike:
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
			// sendstream(stream, groupstore, groupname)

		case *pb.GroupChatRequest_Print:
			fmt.Printf("Need to implement this feature")
			// sendstream(stream, groupstore, groupname)

		default:
			log.Printf("do nothing")
		}

	}
}

// func sendstream(stream pb.ChatService_GroupChatServer, groupstore GroupStore, groupname string) {
// 	group, err := groupstore.GetGroup(groupname)
// 	if err != nil {
// 		log.Printf("error sending group to client %s", err)
// 	}
// 	res := &pb.GroupChatResponse{
// 		Group: group,
// 	}
// 	stream.Send(res)
// 	log.Printf("sending group data of: %s", groupname)
// }

func sendstream(stream pb.ChatService_GroupChatServer, broadcast chan *pb.GroupChatResponse) {
	for {
		res := <-broadcast
		stream.Send(res)
		log.Printf("Stream broadcasted to all..")
	}
}
