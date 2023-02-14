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

func NewAuthServiceClient(authservice pb.AuthServiceClient, store ClientStore) *UserAuthServiceClient {
	return &UserAuthServiceClient{
		authService: authservice,
		clientstore: store,
	}
}

func JoinGroup(groupname string, client *ChatServiceClient) error {
	ctx := context.Background()
	user := client.clientstore.GetUser()
	if user.Name == "None" {
		return fmt.Errorf("Please Login")
	}
	joinchat := &pb.JoinChat{
		Groupname: groupname,
		User:      user,
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

func UserLogin(user_name string, client *UserAuthServiceClient) (*pb.LoginResponse, error) {
	user := &pb.User{
		Id:   uuid.New().ID(),
		Name: user_name,
	}
	user_details := &pb.LoginRequest{
		User: user,
	}
	res, err := client.authService.Login(context.Background(), user_details)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	}
	client.clientstore.SetUser(user)
	log.Printf("User %v Logged in succesfully.", client.clientstore.GetUser().Name)

	return res, nil
}

func GroupChat(client *ChatServiceClient) error {
	//adding grouname to metadata
	md := metadata.Pairs("groupname", client.clientstore.GetGroup().Groupname)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

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

		Print(res.GetGroup())
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
					Chatmessage: &pb.ChatMessage{
						MessagedBy: client.clientstore.GetUser(),
						Message:    msg,
						LikedBy:    make(map[uint32]string, 0),
					},
				},
			}
			req := &pb.GroupChatRequest{
				Action: appendchat,
			}
			stream.Send(req)
			log.Printf("appended a message")

		case "l":
			messagenumber, err := strconv.ParseUint(msg, 10, 32)
			if err != nil {
				log.Printf("please provide a valid number to like")
				continue
			}
			likemessage := &pb.GroupChatRequest_Like{
				Like: &pb.LikeMessage{
					User:      client.clientstore.GetUser(),
					Messageid: uint32(messagenumber),
					Group:     client.clientstore.GetGroup(),
				},
			}
			req := &pb.GroupChatRequest{
				Action: likemessage,
			}
			stream.Send(req)

		case "r":
			messagenumber, err := strconv.ParseUint(msg, 10, 32)
			if err != nil {
				log.Printf("please provide a valid number to unlike")
			}
			unlikemessage := &pb.GroupChatRequest_Unlike{
				Unlike: &pb.UnLikeMessage{
					User:      client.clientstore.GetUser(),
					Messageid: uint32(messagenumber),
					Group:     client.clientstore.GetGroup(),
				},
			}
			req := &pb.GroupChatRequest{
				Action: unlikemessage,
			}
			stream.Send(req)
		
		case "j":
			//join the group
			groupname := strings.TrimSpace(args[1])
			err = JoinGroup(groupname, client)
			if err != nil {
				log.Printf("Failed to join a group: %v", err)
				continue
			}
			//start stream
			log.Printf("starting streaming")
			GroupChat(client)

		default:
			log.Printf("Cannot read the message, please enter again")
			continue
		}
	}
}

func Print(group *pb.Group) {

	groupname := group.GetGroupname()
	participants := maps.Values(group.GetParticipants())
	chatmessages := group.GetMessages()
	chatlength := len(chatmessages)
	print_recent := 10
	fmt.Printf("Group: %v\n", groupname)
	fmt.Printf("Participants: %v\n", participants)
	for print_recent > 0 {
		i := uint32(chatlength - print_recent)

		chatmessage, found := chatmessages[i]
		if found {
			fmt.Printf("%v. %v: %v                     likes: %v\n", i, chatmessage.MessagedBy.Name, chatmessage.Message, len(chatmessage.LikedBy))
		}
		print_recent--
	}

}
