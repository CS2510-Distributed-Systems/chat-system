package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strconv"

	"chat-system/pb"
	"chat-system/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
	chatclient := pb.NewChatServiceClient(conn)
	authclient := pb.NewAuthServiceClient(conn)

	err = readInput(chatclient, authclient)
	if err != nil {
		log.Fatalf("cannot read the input: %s", err)
	}

}

func readInput(client pb.ChatServiceClient, auth pb.AuthServiceClient) error {
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
		case "j":
			groupname := strings.TrimSpace(args[1])
			err = service.JoinGroup(groupname, client)
			if err != nil {
				return status.Errorf(codes.Unavailable, "It faced few errors: %w", err)

			}
		case "u":
			

		default:
			log.Printf("incorrect command, please enter again\n")

		}

	}
}
