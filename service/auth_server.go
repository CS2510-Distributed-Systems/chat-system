package service

import (
	"chat-system/pb"
	"context"
	"log"

	"github.com/google/uuid"
)

type UserAuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *UserAuthServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user_name := req.User.Name
	log.Printf("Logging as: %v", user_name)
	user_details := &pb.User{
		Id:   uuid.New().String(),
		Name: user_name,
	}

	response, err := newUserStore().SaveUser(user_details)
	resp := pb.LoginResponse{
		Id: user_details.Id,
	}
	if err != nil {
		log.Printf("Failed to login user: %v", err)
	}
	log.Println(response.user_auth)
	return &resp, nil
}

func (s *UserAuthServiceServer) NewUserCreate(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user_name := req.User.Name
	log.Printf("Creating new user: %v", user_name)
	user_details := &pb.User{
		Id:   uuid.New().String(),
		Name: user_name,
	}
	//Server_Storage.user_auth[user_details.Id] = user_details
	//user_memory.user_auth[user_details.Id] = user_details
	response := pb.LoginResponse{
		Id: user_details.Id,
	}
	//log.Println(Server_Storage)
	return &response, nil
}
