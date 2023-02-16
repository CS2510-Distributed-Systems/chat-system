# Grpc chat-system


This is a simple console based chat system implemented with Grpc in Go. Grpc uses HTTP2 internally for fast connectivity. Grpc also provides bi-directional streaming, which helps to stream data both ways.

### Usage:

These are the commands that can be used to run the application:
**-make clean**
  This command clean all the files in pb folder which contains proto and grpc proto files.
**-make gen**
  This command generates proto and grpc proto files for the latest code.
**-make server**
  This command starts server whic listens on port 12000.
**-make client**
   This command starts client connection.
 Once both client and server are up and running, User will have to give 0.0.0.0:12000 (any IP address and port must be 12000). This can be done using the following command:
 **-c <ip address>:12000*
  
 Once connected to server user will have to login by giving the following command.
 **-u <username>**
 Now user can perform following actions after logging in:
  **- j <groupname>**: To join a group. If the given group doesn't exist, then a new group will be created and the user will be added to participants list.                        If the user is already present in the group, then that user will be removed from the current group and will be added to the new                            group.
  **- a <message> ** : This command is used to append message to the chat.
  **- l<message_id>**: This command is used to like a message. Here message id is the number displayed before every message.  A user can not like his own                          message.
  **- r<message_id>**: This command is used to dislike a message. A user can only unlike a message iff the user had liked the same message beforehand.
  **- q**            : This command can be used at anytime after runns the client. This command terminates the client session and also removes all                                information related from the server storage.
  
