package service

import (
	"chat-system/pb"
	"context"
	"log"
	//"github.com/google/uuid"
)

type ChatServiceServer struct {
	pb.UnimplementedChatServiceServer
	group_store GroupStore
}
type MessageServiceServer struct {
	pb.UnimplementedMessageServiceServer
	group_store GroupStore
}
type LikeServiceServer struct {
	pb.UnimplementedLikeServiceServer
	group_store GroupStore
}
type UnLikeServiceServer struct {
	pb.UnimplementedUnLikeServiceServer
	group_store GroupStore
}

func NewChatServiceServer(groupstore GroupStore) pb.ChatServiceServer {
	return &ChatServiceServer{
		group_store: groupstore,
	}
}

func NewMessageServiceServer(groupstore GroupStore) pb.MessageServiceServer {
	return &MessageServiceServer{
		group_store: groupstore,
	}
}
func NewLikeServiceServer(groupstore GroupStore) pb.LikeServiceServer {
	return &LikeServiceServer{
		group_store: groupstore,
	}
}
func NewUnLikeServiceServer(groupstore GroupStore) pb.UnLikeServiceServer {
	return &UnLikeServiceServer{
		group_store: groupstore,
	}
}

// rpc
func (s *ChatServiceServer) JoinGroup(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {

	group := req.GetJoinchat()
	log.Printf("receive a group join request with name: %s", group.Groupname)
	CheckUserCurrentGroup(s, req)
	group_details, err := s.group_store.JoinGroup(group.Groupname, group.User)
	//save the group details in the group store
	log.Println(group_details)
	if err != nil {
		log.Printf("Failed to join group %v", err)
	}
	log.Printf("saved laptop with name: %s", group.Groupname)
	res := &pb.JoinResponse{
		Group: group_details,
	}
	return res, nil
}

func CheckUserCurrentGroup(s *ChatServiceServer, req *pb.JoinRequest) {
	user_data := req.Joinchat.GetUser()
	s.group_store.RemoveUserFromCurrentGroup(user_data)
}

func (s *MessageServiceServer) SendMessage(ctx context.Context, req *pb.AppendRequest) (*pb.AppendResponse, error) {
	chat_data := req.GetAppendchat()
	group_data := chat_data.GetGroup()
	message_data := chat_data.GetMessage()
	user_data := chat_data.GetUser()
	message_details := &pb.AppendChat{
		Group:   group_data,
		Message: message_data,
		User:    user_data,
	}
	response, err := s.group_store.AppendMessage(message_details)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
	return response, nil
}

func (s *LikeServiceServer) LikeMessage(ctx context.Context, req *pb.LikeRequest) (*pb.LikeResponse, error) {
	like_data_req := req.GetLikemessage()
	group_data := like_data_req.GetGroup()
	message_id := like_data_req.GetMessageid()
	user_data := like_data_req.GetUser()
	like_data := &pb.LikeMessage{
		Group:     group_data,
		Messageid: message_id,
		User:      user_data,
	}
	err := s.group_store.LikeMessage(like_data)
	return &pb.LikeResponse{Liked: true}, err
}
func (s *UnLikeServiceServer) UnLikeMessage(ctx context.Context, req *pb.LikeRequest) (*pb.LikeResponse, error) {
	unlike_data_req := req.GetLikemessage()
	group_data := unlike_data_req.GetGroup()
	message_id := unlike_data_req.GetMessageid()
	user_data := unlike_data_req.GetUser()
	unlike_data := &pb.LikeMessage{
		Group:     group_data,
		Messageid: message_id,
		User:      user_data,
	}
	err := s.group_store.UnLikeMessage(unlike_data)
	return &pb.LikeResponse{Liked: true}, err
}

func (s *ChatServiceServer) TerminateClientSession(ctx context.Context, req *pb.User) (*pb.TerminateResponse, error) {
	s.group_store.RemoveUserFromCurrentGroup(req)
	return &pb.TerminateResponse{Status: true}, nil
}
