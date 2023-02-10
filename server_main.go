package main


import (
	"fmt"
	"net"
	"strconv"
	"flag"
	"log"
)

const MAX_LINE int = 1024

func main() {
	// Process commandline arguments
	portArg := flag.Int("port", 80, "the port to listen for connections on")

	flag.Parse()

	port := *portArg

	fmt.Println("Welcome to the chat server.")
	connType := "tcp"
	//start server
	conn := startServer(connType, port)

	//start accepting connections from client
	newClient(conn)
	

}

func startServer(connType string, port int)  net.Listener {
	fmt.Println("Listening for data on port", port)
	conn, err := net.Listen(connType, ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	return conn
}


func newClient(conn net.Listener){
	for{
		//accept client
		fmt.Println("new client request recieved")
		new_client_conn, err := conn.Accept()
		if err != nil {
			log.Fatal(err)
		}
		//start serving the clients
		go serveClient(new_client_conn)
	}
}

func serveClient(new_client_conn net.Conn){
	// Receive data from client (store in buf)
	for {
		buf := make([]byte, MAX_LINE)
		ret, err := new_client_conn.Read(buf)
		if err != nil {
			fmt.Printf("error in reading the data from client: %s", err)
		}
		fmt.Printf("received a message from client: %s", buf)
		response  := processData(string(buf[:ret]))
		//send response to connected client
		ret, err = new_client_conn.Write(response)
		if err != nil {
			log.Fatal(err)
		}
	}
	
}

func processData(msgstr string) []byte {
	response_str := []byte(msgstr)
	return response_str
}