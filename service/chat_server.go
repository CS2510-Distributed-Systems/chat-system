package service

import (
	"chat-system/pb"
	"context"
	"log"

	"github.com/google/uuid"
)

type ChatServiceServer struct {
	pb.UnimplementedChatServiceServer
	store ServerStore
}

// rpc
func (s *ChatServiceServer) JoinGroup(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {

	group := req.GetJoinchat()
	log.Printf("receive a group join request with name: %s", group.Groupname)

	//save the group details in the group store
	// id, err := s.store.SaveUser(group.GetUser())
	// if err!= nil{
	// 	return nil, status.Errorf(codes.Internal, "cannot save user to store: %v", err)
	// }

	newgroup := &pb.Group{
		GroupID:   uuid.New().String(),
		Groupname: group.Groupname,

		Participants: []*pb.User{
			{
				Id:   uuid.New().ID(),
				Name: "Dilip",
			},
			{
				Id:   uuid.New().ID(),
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

	log.Printf("saved laptop with name: %s", group.Groupname)
	res := &pb.JoinResponse{
		Group: newgroup,
	}
	return res, nil
}
