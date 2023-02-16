// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: user_service.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AuthServiceClient is the client API for AuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthServiceClient interface {
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, "/chat.AuthService/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService service.
// All implementations should embed UnimplementedAuthServiceServer
// for forward compatibility
type AuthServiceServer interface {
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
}

// UnimplementedAuthServiceServer should be embedded to have forward compatible implementations.
type UnimplementedAuthServiceServer struct {
}

func (UnimplementedAuthServiceServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

// UnsafeAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthServiceServer will
// result in compilation errors.
type UnsafeAuthServiceServer interface {
	mustEmbedUnimplementedAuthServiceServer()
}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

func _AuthService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chat.AuthService/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthService_ServiceDesc is the grpc.ServiceDesc for AuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chat.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Login",
			Handler:    _AuthService_Login_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "user_service.proto",
}

// NewUserCreateClient is the client API for NewUserCreate service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NewUserCreateClient interface {
	NewUserCreate(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
}

type newUserCreateClient struct {
	cc grpc.ClientConnInterface
}

func NewNewUserCreateClient(cc grpc.ClientConnInterface) NewUserCreateClient {
	return &newUserCreateClient{cc}
}

func (c *newUserCreateClient) NewUserCreate(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, "/chat.NewUserCreate/NewUserCreate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NewUserCreateServer is the server API for NewUserCreate service.
// All implementations should embed UnimplementedNewUserCreateServer
// for forward compatibility
type NewUserCreateServer interface {
	NewUserCreate(context.Context, *LoginRequest) (*LoginResponse, error)
}

// UnimplementedNewUserCreateServer should be embedded to have forward compatible implementations.
type UnimplementedNewUserCreateServer struct {
}

func (UnimplementedNewUserCreateServer) NewUserCreate(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewUserCreate not implemented")
}

// UnsafeNewUserCreateServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NewUserCreateServer will
// result in compilation errors.
type UnsafeNewUserCreateServer interface {
	mustEmbedUnimplementedNewUserCreateServer()
}

func RegisterNewUserCreateServer(s grpc.ServiceRegistrar, srv NewUserCreateServer) {
	s.RegisterService(&NewUserCreate_ServiceDesc, srv)
}

func _NewUserCreate_NewUserCreate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NewUserCreateServer).NewUserCreate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chat.NewUserCreate/NewUserCreate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NewUserCreateServer).NewUserCreate(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NewUserCreate_ServiceDesc is the grpc.ServiceDesc for NewUserCreate service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NewUserCreate_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "chat.NewUserCreate",
	HandlerType: (*NewUserCreateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewUserCreate",
			Handler:    _NewUserCreate_NewUserCreate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "user_service.proto",
}