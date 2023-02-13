package service

import (
	"chat-system/pb"
	"context"
	"log"

	"github.com/google/uuid"
)

type UserAuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	store UserStore
	
}

func NewUserAuthServiceServer(userstore UserStore) pb.AuthServiceServer {
	return &UserAuthServiceServer{
		store: userstore,
	}
}


func (s *UserAuthServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user_name := req.User.GetName()
	log.Printf("Logging as: %v", user_name)
	newUser := &pb.User{
		Id:   uuid.New().ID(),
		Name: user_name,
	}
	s.store.SaveUser(newUser)
	res := &pb.LoginResponse{
		User: req.GetUser(),
	}

	return res, nil
}


