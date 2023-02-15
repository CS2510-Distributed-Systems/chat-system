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
	chatservice pb.ChatServiceClient
	authservice pb.AuthServiceClient
	clientstore ClientStore
}

func NewChatServiceClient(chatservice pb.ChatServiceClient, authservice pb.AuthServiceClient, store ClientStore) *ChatServiceClient {
	return &ChatServiceClient{
		chatservice: chatservice,
		authservice: authservice,
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
	res, err := client.chatservice.JoinGroup(ctx, req)
	if err != nil {
		return err
	}
	err = client.clientstore.SetGroup(res.GetGroup())

	fmt.Printf("joined Group %s\n", client.clientstore.GetGroup().Groupname)

	return nil

}

func UserLogin(user_name string, client *ChatServiceClient) error {
	user := &pb.User{
		Id:   uuid.New().ID(),
		Name: user_name,
	}
	user_details := &pb.LoginRequest{
		User: user,
	}

	_, err := client.authservice.Login(context.Background(), user_details)
	if err != nil {
		return fmt.Errorf("Failed to create user: %v", err)
	}
	client.clientstore.SetUser(user)
	log.Printf("User %v Logged in succesfully.", client.clientstore.GetUser().Name)

	return nil
}

func UserLogout(client *ChatServiceClient) bool {
	user := client.clientstore.GetUser()
	req := &pb.LogoutRequest{
		User: &pb.Logout{
			User: user,
		},
	}
	resp, err := client.authservice.Logout(context.Background(), req)
	if err != nil {
		log.Println("cannot Logout.Please Try again")
	}
	client.clientstore.SetUser(
		&pb.User{
			Name: "",
			Id:   0,
		},
	)
	log.Printf("User %v Logged out succesfully.", user.Name)
	return resp.Status
}

func GroupChat(client *ChatServiceClient) error {
	//adding grouname to metadata
	md := metadata.Pairs("groupname", client.clientstore.GetGroup().Groupname)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := client.chatservice.GroupChat(ctx)
	if err != nil {
		return err
	}

	waitResponse := make(chan error)
	// go routine to receive responses
	go receive(stream, waitResponse)

	go send(stream, waitResponse, client)

	
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
		command := res.Command
		if command == "p" {
			log.Println("printing all the messages")
			PrintAll(res.Group)
		} else {
			PrintRecent(res.Group)
		}

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
		//args[1] = strings.Join(args[1:], " ")
		msg = strings.Join(args[1:], " ")
		switch cmd {

		case "a":
			if client.clientstore.GetUser().GetName() == "" {
				log.Println("Please login to join a group.")
			} else if client.clientstore.GetGroup().Groupname == "" {
				log.Println("Please join a group to send a message")
			} else {
				//appending a message
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
				log.Printf("appended a message in the group")
			}

		case "l":
			if client.clientstore.GetUser().GetName() == "" {
				log.Println("Please login to join a group.")
			} else if client.clientstore.GetGroup().Groupname == "" {
				log.Println("Please join a group to send a message")
			} else {
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
			}

		case "r":
			if client.clientstore.GetUser().GetName() == "" {
				log.Println("Please login to join a group.")
			} else if client.clientstore.GetGroup().Groupname == "" {
				log.Println("Please join a group to send a message")
			} else {
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
			}

		case "p":
			{
				print := &pb.GroupChatRequest_Print{
					Print: &pb.PrintChat{
						User:      client.clientstore.GetUser(),
						Groupname: client.clientstore.GetGroup().Groupname,
					},
				}
				req := &pb.GroupChatRequest{
					Action: print,
				}
				stream.Send(req)
			}

		case "j":
			if client.clientstore.GetUser().GetName() == "" {
				log.Println("Please login to join a group.")
			} else if client.clientstore.GetGroup().Groupname == "" {
				log.Println("Please join a group to send a message")
			} else {
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
			}

		case "q":
			logout := &pb.GroupChatRequest_Logout{
				Logout: &pb.Logout{
					User: client.clientstore.GetUser(),
				},
			}
			req := &pb.GroupChatRequest{
				Action: logout,
			}

			stream.Send(req)
			stream.CloseSend()
			return nil

		default:
			log.Printf("Cannot read the message, please enter again")
			continue
		}
	}

}

func PrintRecent(group *pb.Group) {
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
func PrintAll(group *pb.Group) {
	groupname := group.GetGroupname()
	participants := maps.Values(group.GetParticipants())
	chatmessages := group.GetMessages()
	chatlength := len(chatmessages)
	fmt.Printf("Group: %v\n", groupname)
	fmt.Printf("Participants: %v\n", participants)
	count := 0
	for count < chatlength {
		i := uint32(count)

		chatmessage, found := chatmessages[i]
		if found {
			fmt.Printf("%v. %v: %v                     likes: %v\n", i, chatmessage.MessagedBy.Name, chatmessage.Message, len(chatmessage.LikedBy))
		}
		count++
	}

}
