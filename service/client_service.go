package service

import (
	"bufio"
	"chat-system/pb"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/metadata"
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
		Groupname: groupname,
		User: client.clientstore.GetUser(),
	}

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
		Id:   uuid.New().ID(),
		Name: user_name,
	}
	user_details := &pb.LoginRequest{
		User: user,
	}
	res, err := client.Login(context.Background(), user_details)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	}
	log.Printf("User %v Logged in succesfully.", res.User.GetName())
	
	return res, nil
}

func GroupChat(client *ChatServiceClient) error {
	//adding grouname to metadata
	md := metadata.Pairs("groupname",client.clientstore.GetGroup().Groupname)
	ctx := metadata.NewOutgoingContext(context.Background(),md)

	stream, err := client.service.GroupChat(ctx)
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

		Print(res.Group)
	}
}


func send(stream pb.ChatService_GroupChatClient, waitResponse chan error, client *ChatServiceClient) error {
	for {
		log.Printf("Enter the message in the stream:")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatalf("Cannot read the message, please enter again\n")
		}
		input = strings.Trim(input, "\r\n")
		args := strings.Split(input, " ")
		cmd := strings.TrimSpace(args[0])
		msg := strings.TrimSpace(args[1])
		switch cmd {

		case "a":
			appendchat := &pb.GroupChatRequest_Append{
				Append: &pb.AppendChat{
					Group: client.clientstore.GetGroup(),
					Chatmessage : &pb.ChatMessage{
						MessagedBy: client.clientstore.GetUser(),
						Message: msg,
						LikedBy: make(map[uint32]string),
					},
				},
			}
			req := &pb.GroupChatRequest{
				Action: appendchat,
			}
			stream.Send(req)
			log.Printf("appended a message")

		case "l":
			messagenumber,err := strconv.ParseUint(msg,10,32)
			if err !=nil {
				log.Printf("please provide a valid number to like")
			}
			likemessage := &pb.GroupChatRequest_Like{
				Like: &pb.LikeMessage{
					User: client.clientstore.GetUser(),
					Messageid: uint32(messagenumber),
					Group: client.clientstore.GetGroup(),
				},
			}
			req := &pb.GroupChatRequest{
				Action: likemessage,
			}
			stream.Send(req)
			log.Printf("liked a message")

		case "r":
			messagenumber,err := strconv.ParseUint(msg,10,32)
			if err !=nil {
				log.Printf("please provide a valid number to like")
			}
			unlikemessage := &pb.GroupChatRequest_Unlike{
				Unlike: &pb.UnLikeMessage{
					User: client.clientstore.GetUser(),
					Messageid: uint32(messagenumber),
					Group: client.clientstore.GetGroup(),
				},
			}
			req := &pb.GroupChatRequest{
				Action: unlikemessage,
			}
			stream.Send(req)
			log.Printf("liked a message")
		
		default:
			log.Printf("Cannot read the message, please enter again")
			continue
		}
	}
}

func Print(group *pb.Group){
	groupname := group.Groupname
	participants := maps.Values(group.Participants)
	messages := maps.Values(group.Messages)
	fmt.Printf("Group: %v\n",groupname)
	fmt.Printf("Participants: %v\n", participants)
	fmt.Printf("%v\n",messages)
}