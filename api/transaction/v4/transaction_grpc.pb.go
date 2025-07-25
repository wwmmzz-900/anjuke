// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v6.31.1
// source: api/transaction/v4/transaction.proto

package v4

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	Transaction_CreateTransaction_FullMethodName = "/api.transaction.v4.Transaction/CreateTransaction"
)

// TransactionClient is the client API for Transaction service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TransactionClient interface {
	CreateTransaction(ctx context.Context, in *CreateTransactionRequest, opts ...grpc.CallOption) (*CreateTransactionReply, error)
}

type transactionClient struct {
	cc grpc.ClientConnInterface
}

func NewTransactionClient(cc grpc.ClientConnInterface) TransactionClient {
	return &transactionClient{cc}
}

func (c *transactionClient) CreateTransaction(ctx context.Context, in *CreateTransactionRequest, opts ...grpc.CallOption) (*CreateTransactionReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateTransactionReply)
	err := c.cc.Invoke(ctx, Transaction_CreateTransaction_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TransactionServer is the server API for Transaction service.
// All implementations must embed UnimplementedTransactionServer
// for forward compatibility
type TransactionServer interface {
	CreateTransaction(context.Context, *CreateTransactionRequest) (*CreateTransactionReply, error)
	mustEmbedUnimplementedTransactionServer()
}

// UnimplementedTransactionServer must be embedded to have forward compatible implementations.
type UnimplementedTransactionServer struct {
}

func (UnimplementedTransactionServer) CreateTransaction(context.Context, *CreateTransactionRequest) (*CreateTransactionReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTransaction not implemented")
}
func (UnimplementedTransactionServer) mustEmbedUnimplementedTransactionServer() {}

// UnsafeTransactionServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TransactionServer will
// result in compilation errors.
type UnsafeTransactionServer interface {
	mustEmbedUnimplementedTransactionServer()
}

func RegisterTransactionServer(s grpc.ServiceRegistrar, srv TransactionServer) {
	s.RegisterService(&Transaction_ServiceDesc, srv)
}

func _Transaction_CreateTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransactionServer).CreateTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Transaction_CreateTransaction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TransactionServer).CreateTransaction(ctx, req.(*CreateTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Transaction_ServiceDesc is the grpc.ServiceDesc for Transaction service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Transaction_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.transaction.v4.Transaction",
	HandlerType: (*TransactionServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateTransaction",
			Handler:    _Transaction_CreateTransaction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/transaction/v4/transaction.proto",
}
