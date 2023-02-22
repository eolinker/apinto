package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	service "github.com/eolinker/apinto/example/grpc/demo_service"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"

	"golang.org/x/net/context"
)

var _ service.HelloServer = (*Server)(nil)

type Server struct {
	service.UnimplementedHelloServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Hello(ctx context.Context, request *service.HelloRequest) (*service.HelloResponse, error) {

	trailingMD, ok := metadata.FromIncomingContext(ctx)
	if ok {
		grpc.SetTrailer(ctx, trailingMD)
	}
	return &service.HelloResponse{
		Msg: fmt.Sprintf("hello,%s", request.Name),
	}, nil
}

type Request struct {
	Name string
	err  error
}

func (s *Server) StreamRequest(server service.Hello_StreamRequestServer) error {
	trailingMD, ok := metadata.FromIncomingContext(server.Context())
	if ok {
		grpc.SetTrailer(server.Context(), trailingMD)
	}
	msg := make([]string, 0, 10)
	for {
		req, err := server.Recv()
		if err == io.EOF {
			// 开始返回数据
			server.SendMsg(&service.HelloResponse{Msg: strings.Join(msg, "\n")})
			return nil
		}
		if err != nil {
			return nil
		}
		msg = append(msg, req.Name)
	}
}

func (s *Server) StreamResponse(request *service.HelloRequest, server service.Hello_StreamResponseServer) error {
	trailingMD, ok := metadata.FromIncomingContext(server.Context())
	if ok {
		grpc.SetTrailer(server.Context(), trailingMD)
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := server.Send(&service.HelloResponse{
				Msg: fmt.Sprintf("now is %s,name is %s", time.Now().Format("2006-01-02 15:04:05"), request.Name),
			})
			if err != nil && err != io.EOF {
				return err
			}
			return nil
		}
	}
}

func serverStream(server service.Hello_StreamRequestServer, responseServer service.Hello_StreamResponseServer) error {
	ret := make(chan error)
	go func() {
		defer close(ret)
		for {
			req, err := server.Recv()
			if err != nil {
				ret <- err
				return
			}
			err = responseServer.Send(&service.HelloResponse{
				Msg: req.Name,
			})
			if err != nil {
				ret <- err
				return
			}
		}
	}()
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	for {
		select {
		case <-ctx.Done():
			server.SendAndClose(&service.HelloResponse{Msg: "close stream"})
			return nil
		case err, ok := <-ret:
			if !ok {
				return nil
			}
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
}

type AllStreamRequestServer struct {
	service.Hello_AllStreamServer
}

func (s *AllStreamRequestServer) SendAndClose(r *service.HelloResponse) error {
	return s.Send(r)
}

func (s *Server) AllStream(server service.Hello_AllStreamServer) error {
	trailingMD, ok := metadata.FromIncomingContext(server.Context())
	if ok {
		grpc.SetTrailer(server.Context(), trailingMD)
	}
	return serverStream(&AllStreamRequestServer{server}, server)
}
