package service

import (
	"chat-system/pb"
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type chatServiceClient struct {
	service pb.ChatServiceClient
}
type UserAuthServiceClient struct {
	authService pb.AuthServiceClient
}

func NewChatServiceClient(conn *grpc.ClientConn) *chatServiceClient {
	return &chatServiceClient{
		service: pb.NewChatServiceClient(conn),
	}
}

func JoinGroup(groupname string, client pb.ChatServiceClient) error {
	ctx := context.Background()

	joinchat := &pb.JoinChat{
		Groupname: "kothiz",
		User: &pb.User{
			Name: "Dilip",
			Id:   uuid.New().String(),
		}}

	req := &pb.JoinRequest{
		Joinchat: joinchat,
	}
	// fmt.Printf("a group is created in server: %s\n",joinchat )
	res, err := client.JoinGroup(ctx, req)
	if err != nil {
		return err
	}

	fmt.Printf("a group is created in server with groupid: %s", res.GetGroup())

	return nil

}

func (client *chatServiceClient) JoinGroup(ctx context.Context, joinchat *pb.JoinRequest) {

	// // set timeout
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// fmt.Printf("a group is created in server with groupid: ")
	res, err := client.service.JoinGroup(ctx, joinchat)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			// not a big deal
			log.Print("group exists..joining the group")
		} else {
			log.Fatal("cannot create laptop: ", err)
		}
		return
	}

	fmt.Printf("a group is created in server with groupid: %s\n", res.Group.GroupID)

}

func UserLogin(user_name string, client pb.AuthServiceClient) (*pb.LoginResponse, error) {
	user_details := &pb.LoginRequest{
		User: &pb.User{
			Id:   uuid.New().String(),
			Name: user_name,
		},
	}
	res, err := client.Login(context.Background(), user_details)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	}
	log.Printf("User %v Logged in succesfully.",res.GetId())
	return res, nil
}
