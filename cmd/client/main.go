package main

import (
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strconv"
	"chat-system/pb"
)

func main() {
	// Process commandline arguments
	portArg := flag.Int("port", 12000, "the port to listen to the messages on")
	addrArg := flag.String("addr", "localhost", "serveraddr of the server")

	flag.Parse()

	port := *portArg
	serverAddr := *addrArg
	log.Printf("Dialing to server %s:%v", &serverAddr, port)

	// Connect to RPC server
	transportOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(serverAddr+":"+strconv.Itoa(port),transportOption )
	if err != nil {
		log.Fatal("cannot dial the server", err)
	}

	chatclient := pb.NewChatServiceClient(conn)

	



}
