package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

const MAX_LINE int = 1024

func main() {

	// Process commandline arguments
	portArg := flag.Int("port", 12000, "the port to listen to the messages on")
	addrArg := flag.String("addr", "localhost", "serveraddr of the server")

	flag.Parse()

	port := *portArg
	serveraddr := *addrArg

	fmt.Println("Hi All!! Welcome to the chat application")
	connType := "TCP"
	conn := connectServer(serveraddr, port, connType)
	startChat(conn)

}

func connectServer(serveraddr string, port int, connType string) net.Conn {
	// if connType == "UDP" {
	// 	return connectServerUDP(serveraddr, port)
	// } else {
	return connectServerTCP(serveraddr, port)
	// }
}

func connectServerUDP(serveraddr string, port int) {

}

func connectServerTCP(serveraddr string, port int) net.Conn {
	conn, err := net.Dial("tcp", serveraddr+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func startChat(conn net.Conn) {
	for {
		fmt.Println("Enter your message:")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		sendMessage(conn, input)
	}

}

func sendMessage(conn net.Conn, message string) {
	//send the message to the server
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, MAX_LINE)
	ret, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Received %d bytes of Reply:  %s \n", ret, buf)

}
