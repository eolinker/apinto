package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
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
		Msg: "hello",
	}, nil
}

type Request struct {
	Name string
	err  error
}

func (s *Server) StreamRequest(server service.Hello_StreamRequestServer) error {
	requestChan := make(chan *Request)
	return serverRcv(server, requestChan, "streamRequest")
}

func (s *Server) StreamResponse(request *service.HelloRequest, server service.Hello_StreamResponseServer) error {

	requestChan := make(chan *Request)
	ticket := time.NewTicker(1 * time.Second)
	defer ticket.Stop()
	ctx, _ := context.WithTimeout(server.Context(), 10*time.Second)
	go func() {
		for {
			select {
			case <-ticket.C:
				requestChan <- &Request{Name: request.Name}
			case <-ctx.Done():
				return
			}
		}
	}()
	return serverSend(ctx, server, requestChan, "streamResponse")
}

func serverRcv(server service.Hello_StreamRequestServer, requestChan chan *Request, method string) error {
	go func() {
		for {
			req, err := server.Recv()
			if err != nil {
				close(requestChan)
				return
			}
			requestChan <- &Request{
				Name: req.Name,
				err:  nil,
			}
		}
	}()
	ticket := time.NewTicker(3 * time.Second)
	data := map[string]string{}
	for {
		select {
		case <-ticket.C:
			// 超时客户端未关闭，则服务端自行关闭，且将数据返回
			content, _ := json.Marshal(data)
			err := server.SendAndClose(&service.HelloResponse{Msg: string(content)})
			if err != nil {
				return err
			}
			return nil
		case req, ok := <-requestChan:
			{
				if !ok {
					continue
				}
				if req.err != nil {
					return req.err
				}

				trailingMD, ok := metadata.FromIncomingContext(server.Context())
				if ok {
					server.SetTrailer(trailingMD)
				}
			}
		}
	}
}

func serverSend(ctx context.Context, server service.Hello_StreamResponseServer, requestChan chan *Request, method string) error {
	newCtx, _ := context.WithTimeout(ctx, 3*time.Second)
	for {
		select {
		case req, ok := <-requestChan:
			{
				if !ok {
					continue
				}
				if req.err != nil {
					return req.err
				}
				now := time.Now().Format("2006-01-02 15:04:05")

				trailingMD, ok := metadata.FromIncomingContext(server.Context())
				if ok {
					server.SetTrailer(trailingMD)
				}
				err := server.Send(&service.HelloResponse{Msg: fmt.Sprintf("Welcome!Now time is %s", now)})
				if err != nil {
					if err != io.EOF {
						return err
					}
					return errors.New("close stream")
				}
			}
		case <-newCtx.Done():
			{
				return nil
			}
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
	requestChan := make(chan *Request)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := serverSend(server.Context(), server, requestChan, "allStream")
		if err != nil {
			log.Println(err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := serverRcv(&AllStreamRequestServer{server}, requestChan, "allStream")
		if err != nil {
			log.Println(err)
			return
		}
	}()
	wg.Wait()
	return nil
}
