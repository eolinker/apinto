package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// interceptor 客户端拦截器
func interceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	fmt.Printf("method=%s req=%v reply=%v duration=%s error=%v\n", method, req, reply, time.Since(start), err)
	return err
}

// streamInterceptor 流客户端拦截器
func streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	start := time.Now()
	client, err := streamer(ctx, desc, cc, method, opts...)
	fmt.Printf("stream_name=%s method=%s is_server_stream=%t is_client_stream=%t duration=%s error=%v\n", desc.StreamName, method, desc.ServerStreams, desc.ClientStreams, time.Since(start), err)
	return client, err
}
