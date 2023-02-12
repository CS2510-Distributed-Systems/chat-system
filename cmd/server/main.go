package main

import (
	"chat-system/pb"
	"chat-system/service"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
)

func main() {
	// Process commandline arguments
	portArg := flag.Int("port", 80, "the server port")
	flag.Parse()
	port := *portArg

	grpcserver := grpc.NewServer()
	groupstore := service.NewInMemoryGroupStore()
	chatserver := service.NewChatServiceServer(groupstore)
	userstore := service.NewInMemoryUserStore()
	authserver := service.NewUserAuthServiceServer(userstore)
	pb.RegisterChatServiceServer(grpcserver, chatserver)
	pb.RegisterAuthServiceServer(grpcserver, authserver)
	log.Printf("start server on port: %d", port)
	Listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("cannot start server: %w", err)
	}
	log.Printf("Start GRPC server at %s", Listener.Addr())

	err = grpcserver.Serve(Listener)
	fmt.Println(err)
	if err != nil {
		return
	}

}
