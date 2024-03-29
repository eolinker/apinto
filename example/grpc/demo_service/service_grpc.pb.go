// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: service.proto

package demo_service

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

// HelloClient is the client API for Hello service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HelloClient interface {
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
	StreamRequest(ctx context.Context, opts ...grpc.CallOption) (Hello_StreamRequestClient, error)
	StreamResponse(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (Hello_StreamResponseClient, error)
	AllStream(ctx context.Context, opts ...grpc.CallOption) (Hello_AllStreamClient, error)
}

type helloClient struct {
	cc grpc.ClientConnInterface
}

func NewHelloClient(cc grpc.ClientConnInterface) HelloClient {
	return &helloClient{cc}
}

func (c *helloClient) Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	out := new(HelloResponse)
	err := c.cc.Invoke(ctx, "/Service.Hello/Hello", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helloClient) StreamRequest(ctx context.Context, opts ...grpc.CallOption) (Hello_StreamRequestClient, error) {
	stream, err := c.cc.NewStream(ctx, &Hello_ServiceDesc.Streams[0], "/Service.Hello/StreamRequest", opts...)
	if err != nil {
		return nil, err
	}
	x := &helloStreamRequestClient{stream}
	return x, nil
}

type Hello_StreamRequestClient interface {
	Send(*HelloRequest) error
	CloseAndRecv() (*HelloResponse, error)
	grpc.ClientStream
}

type helloStreamRequestClient struct {
	grpc.ClientStream
}

func (x *helloStreamRequestClient) Send(m *HelloRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *helloStreamRequestClient) CloseAndRecv() (*HelloResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(HelloResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *helloClient) StreamResponse(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (Hello_StreamResponseClient, error) {
	stream, err := c.cc.NewStream(ctx, &Hello_ServiceDesc.Streams[1], "/Service.Hello/StreamResponse", opts...)
	if err != nil {
		return nil, err
	}
	x := &helloStreamResponseClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Hello_StreamResponseClient interface {
	Recv() (*HelloResponse, error)
	grpc.ClientStream
}

type helloStreamResponseClient struct {
	grpc.ClientStream
}

func (x *helloStreamResponseClient) Recv() (*HelloResponse, error) {
	m := new(HelloResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *helloClient) AllStream(ctx context.Context, opts ...grpc.CallOption) (Hello_AllStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Hello_ServiceDesc.Streams[2], "/Service.Hello/AllStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &helloAllStreamClient{stream}
	return x, nil
}

type Hello_AllStreamClient interface {
	Send(*HelloRequest) error
	Recv() (*HelloResponse, error)
	grpc.ClientStream
}

type helloAllStreamClient struct {
	grpc.ClientStream
}

func (x *helloAllStreamClient) Send(m *HelloRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *helloAllStreamClient) Recv() (*HelloResponse, error) {
	m := new(HelloResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// HelloServer is the server API for Hello service.
// All implementations must embed UnimplementedHelloServer
// for forward compatibility
type HelloServer interface {
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	StreamRequest(Hello_StreamRequestServer) error
	StreamResponse(*HelloRequest, Hello_StreamResponseServer) error
	AllStream(Hello_AllStreamServer) error
	mustEmbedUnimplementedHelloServer()
}

// UnimplementedHelloServer must be embedded to have forward compatible implementations.
type UnimplementedHelloServer struct {
}

func (UnimplementedHelloServer) Hello(context.Context, *HelloRequest) (*HelloResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Hello not implemented")
}
func (UnimplementedHelloServer) StreamRequest(Hello_StreamRequestServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamRequest not implemented")
}
func (UnimplementedHelloServer) StreamResponse(*HelloRequest, Hello_StreamResponseServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamResponse not implemented")
}
func (UnimplementedHelloServer) AllStream(Hello_AllStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method AllStream not implemented")
}
func (UnimplementedHelloServer) mustEmbedUnimplementedHelloServer() {}

// UnsafeHelloServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HelloServer will
// result in compilation errors.
type UnsafeHelloServer interface {
	mustEmbedUnimplementedHelloServer()
}

func RegisterHelloServer(s grpc.ServiceRegistrar, srv HelloServer) {
	s.RegisterService(&Hello_ServiceDesc, srv)
}

func _Hello_Hello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServer).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Service.Hello/Hello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServer).Hello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Hello_StreamRequest_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(HelloServer).StreamRequest(&helloStreamRequestServer{stream})
}

type Hello_StreamRequestServer interface {
	SendAndClose(*HelloResponse) error
	Recv() (*HelloRequest, error)
	grpc.ServerStream
}

type helloStreamRequestServer struct {
	grpc.ServerStream
}

func (x *helloStreamRequestServer) SendAndClose(m *HelloResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *helloStreamRequestServer) Recv() (*HelloRequest, error) {
	m := new(HelloRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Hello_StreamResponse_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(HelloRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HelloServer).StreamResponse(m, &helloStreamResponseServer{stream})
}

type Hello_StreamResponseServer interface {
	Send(*HelloResponse) error
	grpc.ServerStream
}

type helloStreamResponseServer struct {
	grpc.ServerStream
}

func (x *helloStreamResponseServer) Send(m *HelloResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Hello_AllStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(HelloServer).AllStream(&helloAllStreamServer{stream})
}

type Hello_AllStreamServer interface {
	Send(*HelloResponse) error
	Recv() (*HelloRequest, error)
	grpc.ServerStream
}

type helloAllStreamServer struct {
	grpc.ServerStream
}

func (x *helloAllStreamServer) Send(m *HelloResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *helloAllStreamServer) Recv() (*HelloRequest, error) {
	m := new(HelloRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Hello_ServiceDesc is the grpc.ServiceDesc for Hello service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Hello_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Service.Hello",
	HandlerType: (*HelloServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hello",
			Handler:    _Hello_Hello_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamRequest",
			Handler:       _Hello_StreamRequest_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "StreamResponse",
			Handler:       _Hello_StreamResponse_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "AllStream",
			Handler:       _Hello_AllStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "service.proto",
}
