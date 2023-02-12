package service

import (
	"bufio"
	"chat-system/pb"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	// grpc "google.golang.org/grpc"
	// codes "google.golang.org/grpc/codes"
	// status "google.golang.org/grpc/status"
)

type ChatServiceClient struct {
	service     pb.ChatServiceClient
	clientstore ClientStore
}
type UserAuthServiceClient struct {
	authService pb.AuthServiceClient
	clientstore ClientStore
}

func NewChatServiceClient(chatservice pb.ChatServiceClient, store ClientStore) *ChatServiceClient {
	return &ChatServiceClient{
		service:     chatservice,
		clientstore: store,
	}
}

func JoinGroup(groupname string, client *ChatServiceClient) error {
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
	res, err := client.service.JoinGroup(ctx, req)
	if err != nil {
		return err
	}
	err = client.clientstore.SetGroup(res.GetGroup())

	fmt.Printf("joined Group %s\n", client.clientstore.GetGroup().Groupname)

	return nil

}

func UserLogin(user_name string, client pb.AuthServiceClient) (*pb.LoginResponse, error) {
	user := &pb.User{
		Id:   uuid.New().String(),
		Name: user_name,
	}
	user_details := &pb.LoginRequest{
		User: user,
	}
	res, err := client.Login(context.Background(), user_details)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	}
	log.Printf("User %v Logged in succesfully.", res.GetId())
	
	return res, nil
}

func GroupChat(client *ChatServiceClient) error {

	ctx := context.Background()

	stream, err := client.service.GroupChat(ctx, client.clientstore.GetGroup().GetGroupname)
	if err != nil {
		return err
	}

	waitResponse := make(chan error)
	// go routine to receive responses
	go receive(stream, waitResponse)

	go send(stream, waitResponse, client)

	err = <-waitResponse
	return nil
}

func receive(stream pb.ChatService_GroupChatClient, waitResponse chan error) error {
	for {
		fmt.Println("trying to receive from the server continously")
		res, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more responses")
			waitResponse <- nil
			return err
		}
		if err != nil {
			waitResponse <- fmt.Errorf("cannot receive stream response: %v", err)
			return err
		}

		log.Print("received response: ", res)
	}
}

func send(stream pb.ChatService_GroupChatClient, waitResponse chan error, client *ChatServiceClient) error {
	for {
		log.Printf("Enter the message in the stream:")
		msg, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatalf("Cannot read the message, please enter again\n")
		}
		msg = strings.Trim(msg, "\r\n")
		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])
		switch cmd {
		case "a":
			appendchat := &pb.GroupChatRequest_Append{
				Append: &pb.AppendChat{
					User:  client.clientstore.GetUser(),
					Group: client.clientstore.GetGroup(),
					Message: &pb.ChatMessage{
						Text:  strings.TrimSpace(args[1]),
						Likes: 0,
					},
				},
			}
			req := &pb.GroupChatRequest{
				Action: appendchat,
			}
			res := stream.Send(req)
			log.Print("received response after appending: ", res)
		default:
			log.Printf("Cannot read the message, please enter again\n")
			continue
		}
	}
}
