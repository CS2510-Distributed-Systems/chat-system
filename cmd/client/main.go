package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strconv"

	"chat-system/pb"
	"chat-system/service"
	"fmt"

	"google.golang.org/grpc"

	// "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	// "google.golang.org/grpc/status"
	"strings"
)

func main() {
	// Process commandline arguments
	addrArg := flag.String("addr", "localhost", "serveraddr of the server")
	portArg := flag.Int("port", 12000, "the server port")

	flag.Parse()

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
	clientstore := service.NewInMemoryClientStore()
	chatclient := service.NewChatServiceClient(pb.NewChatServiceClient(conn), clientstore)
	authclient := pb.NewAuthServiceClient(conn)

	_, err = readInput(chatclient, authclient)
	conn.Close()

}

func readInput(chatclient *service.ChatServiceClient, authclient pb.AuthServiceClient) (uint32, error) {
	for {
		log.Printf("Enter the message:")
		msg, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatalf("Cannot read the message, please enter again\n")
		}
		msg = strings.Trim(msg, "\r\n")
		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])

		switch cmd {
		case "q":
			//closes the program
			fmt.Println("close the program")
			return 1, nil

		case "u":
			//user login
			username := strings.TrimSpace(args[1])
			_, err := service.UserLogin(username, authclient)
			if err != nil {
				log.Printf("Failed to login: %v", err)

			}
		case "j":
			//join the group
			groupname := strings.TrimSpace(args[1])
			err = service.JoinGroup(groupname, chatclient)
			if err != nil {
				log.Printf("Failed to create a group: %v", err)
			}
			//start stream
			log.Printf("starting streaming")
			service.GroupChat(chatclient)
		default:
			log.Printf("incorrect command, please enter again\n")

		}

	}

}
