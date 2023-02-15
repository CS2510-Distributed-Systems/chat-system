package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"chat-system/pb"
	"chat-system/service"
	"strings"

	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"
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
	chatclient := service.NewChatServiceClient(pb.NewChatServiceClient(conn), pb.NewAuthServiceClient(conn),clientstore)

	for {
		//read input
		log.Printf("Enter the message:")
		msg, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatalf("Cannot read the message, please enter again\n")
		}
		msg = strings.Trim(msg, "\r\n")

		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])
		//args[1] = strings.Join(args[1:], " ")
		arg := strings.Join(args[1:], " ")

		switch cmd {
		case "u":
			username := strings.TrimSpace(arg)
			err := service.UserLogin(username, chatclient)
			if err != nil{
				fmt.Println(err)
				return
			}

		case "j":
			if clientstore.GetUser().GetName() == "" {
				log.Println("Please login to join a group.")
			} else {
				//join the group
				groupname := strings.TrimSpace(arg)
				service.JoinGroup(groupname, chatclient)
				//start streaming
				service.GroupChat(chatclient)
			}
		case "a", "l", "r":
			log.Println("please Enter the chat room")

		case "q":
			if clientstore.GetUser().Name != "" {
				resp := service.UserLogout(chatclient)
				if resp{
					conn.Close()
					log.Println("close the program")
					return
				}else {
					log.Println("Failed to exit program. Please try again.")
				}
			} else {
				conn.Close()
				log.Println("close the program")
				return
			}

		default:
			//helpcode//
			log.Printf("incorrect command, please enter again\n")
		}
	}

}
