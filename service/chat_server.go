package service

import (
	"chat-system/pb"
	"context"
	"fmt"
	"log"
	"time"
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

	//save the group details in the group store
	group_details, err := s.groupstore.JoinGroup(group.Groupname, group.User)
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
	if !ok{
		log.Printf("didnt receive the context properly from the client..")
	}
	groupname := md.Get("groupname")[0]

	errch := make(chan error)
	go receivestream(stream, s.groupstore)
	go sendstream(stream,s.groupstore, groupname)

	return <-errch
}

func receivestream(stream pb.ChatService_GroupChatServer, groupstore GroupStore) {
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("error in receiving")
			break
		}
		switch req.GetAction().(type) {
		case *pb.GroupChatRequest_Append:
			group := req.GetAppend().Group
			message := req.GetAppend().Message
			user := req.GetAppend().User
			message_details := &pb.AppendChat{
				Group:   group,
				Message: message,
				User:    user,
			}
			err := groupstore.AppendMessage(message_details)
			if err != nil {
				log.Printf("some error occured in appending the message: %s", err)
			}
			fmt.Printf("trying to append new line : %s in group : %s", message, group.GetGroupname())

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
				log.Printf("some error occured in liking the message: %s", err)
			}
			fmt.Printf("trying to like line : %s in group : %s", msgId, group.GetGroupname())

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
			fmt.Printf("trying to like line : %s in group : %s", msgId, group.GetGroupname())

		case *pb.GroupChatRequest_Print:			
			fmt.Printf("Need to implement this feature")

		default:
			log.Printf("do nothing")
		}

	}
}

func sendstream(stream pb.ChatService_GroupChatServer, groupstore GroupStore ,groupname string) {
	for {
		time.Sleep(5 * time.Second)
		
		group, err := groupstore.GetGroup(groupname)
		if err != nil {
			log.Printf("error sending group to client %s", err)
		}
		res := &pb.GroupChatResponse{
			Group:group,
		}
		stream.Send(res)
		log.Printf("sending group data of: %s", groupname)
	}
}
