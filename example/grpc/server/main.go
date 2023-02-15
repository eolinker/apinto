package main

import (
	"fmt"
	"net"

	service "github.com/eolinker/apinto/example/grpc/demo_service"

	"google.golang.org/grpc/credentials"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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

func grpcGateway() {
	runtime.NewServeMux()
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
			panic(err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	if err != nil {
		grpclog.Fatalf("Failed to generate credentials %v", err)
	}

	// 实例化grpc Server
	s := grpc.NewServer(opts...)
	ser := NewServer()
	// 注册HelloService
	service.RegisterHelloServer(s, ser)
	//reflection.Register(s)
	fmt.Println("Listen on " + address)

	return s.Serve(l)
}
