package service

import (
	"chat-system/pb"
	"context"
	"fmt"

	
	"log"
	"time"

	"github.com/google/uuid"
)

type ChatServiceServer struct {
	pb.UnimplementedChatServiceServer
}

// rpc
func (s *ChatServiceServer) JoinGroup(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {

	group := req.GetJoinchat()
	log.Printf("receive a group join request with name: %s", group.Groupname)

	//save the group details in the group store

	newgroup := &pb.Group{
		GroupID:   uuid.New().String(),
		Groupname: group.Groupname,

		Participants: []*pb.User{
			{
				Id:   uuid.New().String(),
				Name: "Dilip",
			},
			{
				Id:   uuid.New().String(),
				Name: "Teja",
			},
		},
		Messages: []*pb.ChatMessage{
			{
				Text:  "Hi this is a new group",
				Likes: 0,
			},
			{
				Text:  "Hi this is a again new group",
				Likes: 1,
			},
		},
	}

	log.Printf("saved group with name: %s", group.Groupname)
	res := &pb.JoinResponse{
		Group: newgroup,
	}
	return res, nil
}

//streaming rpc
func (s *ChatServiceServer) GroupChat(stream pb.ChatService_GroupChatServer) error {

	errch := make(chan error)
	go receivestream(stream)
	go sendstream(stream)

	return <-errch
}

func receivestream(stream pb.ChatService_GroupChatServer ) {
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("error in receiving")
			break
		}
		switch req.GetAction().(type){
		case *pb.GroupChatRequest_Append:
			newgroup := &pb.Group{
				GroupID:   uuid.New().String(),
				Groupname: req.GetAppend().String(),

				Participants: []*pb.User{
					{
						Id:   uuid.New().String(),
						Name: "Dilip",
					},
					{
						Id:   uuid.New().String(),
						Name: "Teja",
					},
				},
				Messages: []*pb.ChatMessage{
					{
						Text:  "Hi this is a new group",
						Likes: 0,
					},
					{
						Text:  "Hi this is a again new group",
						Likes: 1,
					},
				},
			}

			fmt.Printf("trying to append new line : %s", newgroup)
			res := &pb.GroupChatResponse{
				Group: newgroup,
			}
			err  = stream.Send(res)
		case *pb.GroupChatRequest_Like:
		case *pb.GroupChatRequest_Unlike:
		case *pb.GroupChatRequest_Print:
		default:
			log.Printf("do nothing")
		}

	}
}

func sendstream(stream pb.ChatService_GroupChatServer) {
	for {
		time.Sleep(5 * time.Second)
		newgroup := &pb.Group{
			GroupID:   uuid.New().String(),
			Groupname: "kothizzz",
			Participants: []*pb.User{
				{
					Id:   uuid.New().String(),
					Name: "Dilip",
				},
				{
					Id:   uuid.New().String(),
					Name: "Teja",
				},
			},
			Messages: []*pb.ChatMessage{
				{
					Text:  "Hi this exactly how i want",
					Likes: 0,
				},
				{
					Text:  "Hi this is a again new group",
					Likes: 1,
				},
			},
		}

		res := &pb.GroupChatResponse{
			Group: newgroup,
		}
		stream.Send(res)
	}
}
