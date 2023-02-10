package main

import (
	"flag"
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


	log.Printf("start server on port: %d", port)
	Listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil{
		log.Fatal("cannot start server: %w",err)
	}

	err = grpcserver.Serve(Listener)
	if err!=nil{
		log.Fatal("Cannot start server", err)
	}




}
