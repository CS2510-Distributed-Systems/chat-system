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
		Id:   uuid.New().String(),
		Name: user_name,
	}
	s.store.SaveUser(newUser)
	res := &pb.LoginResponse{
		User: newUser,
	}

	return res, nil
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
		User: user_details,
	}
	//log.Println(Server_Storage)
	return &response, nil
}
