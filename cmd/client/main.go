package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strconv"

	"chat-system/pb"
	"chat-system/service"
	"strings"

	"google.golang.org/grpc"
	//"google.golang.org/protobuf/internal/encoding/text"
	//"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	//"google.golang.org/grpc/status"
)

func main() {
	// Process commandline arguments
	addrArg := flag.String("addr", "localhost", "serveraddr of the server")
	portArg := flag.Int("port", 12000, "the server port")

	flag.Parse()
	client_details := *&pb.User{
		Id:   "",
		Name: "",
	}
	current_group_details := &pb.Group{
		GroupID:      "",
		Groupname:    "",
		Participants: make(map[string]string),
		Messages:     make(map[string]*pb.MessageDetails),
	}
	port := *portArg
	serverAddr := *addrArg
	log.Printf("Dialing to server %s:%v", serverAddr, port)

	// Connect to RPC server
	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(serverAddr+":"+strconv.Itoa(port), transportOption)
	if err != nil {
		log.Fatal("cannot dial the server", err)
	}

	log.Printf("Dialing to server %s:%v", serverAddr, port)
	// defer conn.Close()
	chatclient := pb.NewChatServiceClient(conn)
	authclient := pb.NewAuthServiceClient(conn)
	messageclient := pb.NewMessageServiceClient(conn)
	likeclient := pb.NewLikeServiceClient(conn)
	unlikeclient := pb.NewUnLikeServiceClient(conn)

	for {
		log.Printf("Enter the message:")
		msg, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatalf("Cannot read the message, please enter again\n")
		}

		msg = strings.Trim(msg, "\r\n")

		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])
		//args[1] = strings.Join(args[1:], " ")
		argument := strings.Join(args[1:], " ")

		switch cmd {
		case "u":
			if client_details.Name != "" {
				log.Println("Please logout from current user session and try again.")
			} else {
				username := strings.TrimSpace(argument)
				resp, err := service.UserLogin(username, authclient)
				client_details.Id = resp.User.Id
				client_details.Name = resp.User.Name
				log.Printf("From Console. Current client = %v", client_details.Name)
				if err != nil {
					log.Printf("Failed to login: %v", err)
				}
			}
		case "j":
			if client_details.Name == "" {
				log.Println("Please login first to join a group.")
			} else {
				groupname := strings.TrimSpace(argument)
				user_data := &pb.User{
					Id:   client_details.Id,
					Name: client_details.Name,
				}
				resp, err := service.JoinGroup(groupname, *user_data, chatclient)
				if err != nil {
					log.Printf("Failed to join group: %v", err)
				}
				current_group_details.GroupID = resp.Group.Groupname
				current_group_details.Groupname = resp.Group.Groupname
				current_group_details.Participants = resp.Group.Participants
				current_group_details.Messages = resp.Group.Messages
			}

		case "a":
			if client_details.Name == "" {
				log.Println("Please login first to join a group.")
			} else if current_group_details.Groupname == "" {
				log.Println("Please join a group to send a message")
			} else {
				message := strings.TrimSpace(argument)
				chat_message := &pb.ChatMessage{
					Text:  message,
					Likes: make(map[string]string),
				}
				message_data := &pb.AppendChat{
					User:    &client_details,
					Group:   current_group_details,
					Message: chat_message,
				}
				append_request := &pb.AppendRequest{
					Appendchat: message_data,
				}
				resp, err := service.SendMessage(append_request, messageclient)
				if err != nil {
					log.Printf("Failed to send message: %v", err)
				}
				print(resp)
			}

		case "l":
			if client_details.Name == "" {
				log.Println("Please login first to join a group.")
			} else if current_group_details.Groupname == "" {
				log.Println("Please join a group to send a message")
			} else {
				messageid := strings.TrimSpace(argument)
				like_message_data := &pb.LikeMessage{
					User:      &client_details,
					Group:     current_group_details,
					Messageid: messageid,
				}
				like_data := &pb.LikeRequest{
					Likemessage: like_message_data,
				}
				resp, err := service.LikeMessage(like_data, likeclient)
				if err != nil {
					log.Printf("Failed to like message: %v", err)
				}
				print(resp)
			}

		case "r":
			if client_details.Name == "" {
				log.Println("Please login first to join a group.")
			} else if current_group_details.Groupname == "" {
				log.Println("Please join a group to send a message")
			} else {
				messageid := strings.TrimSpace(argument)
				unlike_message_data := &pb.LikeMessage{
					User:      &client_details,
					Group:     current_group_details,
					Messageid: messageid,
				}
				unlike_data := &pb.LikeRequest{
					Likemessage: unlike_message_data,
				}
				resp, err := service.UnLikeMessage(unlike_data, unlikeclient)
				if err != nil {
					log.Printf("Failed to unlike message: %v", err)
				}
				print(resp)
			}

		case "q":
			if client_details.Name != "" {
				resp := service.TerminateClientSession(&client_details, chatclient)
				if resp.Status {
					conn.Close()
					log.Fatal("Exited program.")
				} else {
					log.Println("Failed to exit program. Please try again.")
				}
			} else {
				log.Println("You aren't logged into any program.")
			}

		default:
			log.Printf("incorrect command, please enter again\n")
		}

	}

}

// func readInput(client pb.ChatServiceClient) error {
// 	for {
// 		log.Printf("Enter the message:")
// 		msg, err := bufio.NewReader(os.Stdin).ReadString('\n')
// 		if err != nil {
// 			log.Fatalf("Cannot read the message, please enter again\n")
// 		}

// 		msg = strings.Trim(msg, "\r\n")

// 		args := strings.Split(msg, " ")
// 		cmd := strings.TrimSpace(args[0])

// 		switch cmd {
// 		case "j":
// 			groupname := strings.TrimSpace(args[1])
// 			err = service.JoinGroup(groupname,, client)
// 			if err != nil {
// 				return status.Errorf(codes.Unavailable, "It faced few errors: %w", err)

// 			}
// 		default:
// 			log.Printf("incorrect command, please enter again\n")

// 		}

// 	}
// }
