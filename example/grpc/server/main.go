package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc/reflection"

	service "github.com/eolinker/apinto/example/grpc/demo_service"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func main() {
	Parse()
	err := listen()
	if err != nil {
		fmt.Println(err)
		return
	}

}

func listen() error {
	address := fmt.Sprintf("%s:%d", BindIP, ListenPort)
	l, err := net.Listen("tcp", address)
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	if TlsKey != "" && TlsPem != "" {
		creds, err := credentials.NewServerTLSFromFile(TlsPem, TlsKey)
		if err != nil {
			grpclog.Fatalf("Failed to generate credentials %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// 实例化grpc Server
	s := grpc.NewServer(opts...)
	ser := NewServer()
	// 注册HelloService
	service.RegisterHelloServer(s, ser)
	// 开启grpc反射
	reflection.Register(s)
	fmt.Println("Listen on " + address)

	return s.Serve(l)
}
