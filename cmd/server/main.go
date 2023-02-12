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
	portArg := flag.Int("port", 12000, "the server port")
	flag.Parse()
	port := *portArg

	grpcserver := grpc.NewServer()
	groupstore := service.NewInMemoryGroupStore()
	chatserver := service.NewChatServiceServer(groupstore)
	messageserver := service.NewMessageServiceServer(groupstore)
	likeserver := service.NewLikeServiceServer(groupstore)
	unlikeserver := service.NewUnLikeServiceServer(groupstore)
	userstore := service.NewInMemoryUserStore()
	authserver := service.NewUserAuthServiceServer(userstore)
	pb.RegisterChatServiceServer(grpcserver, chatserver)
	pb.RegisterAuthServiceServer(grpcserver, authserver)
	pb.RegisterMessageServiceServer(grpcserver, messageserver)
	pb.RegisterLikeServiceServer(grpcserver, likeserver)
	pb.RegisterUnLikeServiceServer(grpcserver, unlikeserver)
	log.Printf("start server on port: %d", port)
	Listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
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
