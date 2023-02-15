package main

import (
	"fmt"
	"liujian-test/grpc-test-demo/common/flag"
	"liujian-test/grpc-test-demo/server"
	"liujian-test/grpc-test-demo/service"
	"net"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func main() {
	flag.Parse()
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
	address := fmt.Sprintf("%s:%d", flag.BindIP, flag.ListenPort)
	l, err := net.Listen("tcp", address)
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	if flag.TlsKey != "" && flag.TlsPem != "" {
		//cert, err := tls.LoadX509KeyPair(flag.TlsPem, flag.TlsKey)
		//if err != nil {
		//	log.Println(err)
		//}
		//l = tls.NewListener(l, &tls.Config{Certificates: []tls.Certificate{cert}})
		//creds, err := credentials.NewServerTLSFromFile(flag.TlsPem, flag.TlsKey)
		//if err != nil {
		//	panic(err)
		//}
		//opts = append(opts, grpc.Creds(creds))
	}

	if err != nil {
		grpclog.Fatalf("Failed to generate credentials %v", err)
	}

	//注册interceptor
	opts = append(opts,
		grpc.UnaryInterceptor(UnaryServerAuthInterceptor(initAuth())),
		//grpc.StreamInterceptor(StreamServerAuthInterceptor(initAuth())),
	)
	// 实例化grpc Server
	s := grpc.NewServer(opts...)
	ser := server.NewServer()
	// 注册HelloService
	service.RegisterHelloServer(s, ser)
	//reflection.Register(s)
	fmt.Println("Listen on " + address)

	return s.Serve(l)
}
