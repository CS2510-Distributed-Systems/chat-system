# Grpc chat-system


This is a simple console based chat system implemented with Grpc in Go. Grpc uses HTTP2 internally for fast connectivity. Grpc also <br />provides bi-directional  streaming, which helps to stream data both ways.

### Usage:

These are the commands that can be used to run the application:<br />
**-make clean**<br />
 &emsp;  This command clean all the files in pb folder which contains proto and grpc proto files.<br />
**-make gen** <br />
 &emsp;  This command generates proto and grpc proto files for the latest code.<br />
**-make server** <br />
 &emsp;  This command starts server whic listens on port 12000.<br />
**-make client** <br />
 &emsp;   This command starts client connection.<br />
 <br />
Once both client and server are up and running, User will have to give 0.0.0.0:12000 (any IP address and port must be 12000). This can be <br /> done using the following command:<br />
 **-c&nbsp;<ip_address>:12000**<br />
  
 Once connected to server user will have to login by giving the following command.<br />
 **-u&nbsp;<user_name>**<br />
 <br />
 Now user can perform following actions after logging in:<br />
  **- j&nbsp;<group_name>**: <br />&emsp;To join a group. If the given group doesn't exist, then a new group will be created and the user will be added to participants list.<br />If the user is already present in the group, then that user will be removed from the current group and will be added to the new group.<br />
  **- a&nbsp;<message_>** :<br />&emsp; This command is used to append message to the chat.<br />
  **- l&nbsp;<message_id>**:<br /> &emsp;This command is used to like a message. Here message id is the number displayed before every message.  A user can not like his own message.<br />
  **- r&nbsp;<message_id>**:<br /> &emsp;This command is used to dislike a message. A user can only unlike a message iff the user had liked the same message beforehand.<br />
  **-p**             :<br /> &emsp; This command prints all the messages right from the group creation with latest message on bottom and oldest message on top. <br />
  **- q**            :<br /> &emsp;This command can be used at anytime after runs the client. This command terminates the client session and also removes<br /> all information related from the server storage.<br />

Once after joining a group the user user is show with latest 10 messages. User can use **p** command to see all the messages.<br /> For every action performed by the user, all other active participants are broadcasted with the change. 
