package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"liujian-test/grpc-test-demo/service"
	"log"
	"sync"
	"time"

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

	data, md, err := retrieveData("hello", request.Name)
	if err != nil {
		return nil, err
	}
	if md != nil {
		grpc.SendHeader(ctx, md)
	}
	trailingMD, ok := metadata.FromIncomingContext(ctx)
	if ok {
		grpc.SetTrailer(ctx, trailingMD)
	}
	return &service.HelloResponse{
		Msg: data,
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
				tmp, md, err := retrieveData(method, req.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if md != nil {
					server.SendHeader(md)
				}
				trailingMD, ok := metadata.FromIncomingContext(server.Context())
				if ok {
					server.SetTrailer(trailingMD)
				}
				data[req.Name] = tmp
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
				data, md, err := retrieveData(method, req.Name)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if md != nil {
					server.SendHeader(md)
				}
				trailingMD, ok := metadata.FromIncomingContext(server.Context())
				if ok {
					server.SetTrailer(trailingMD)
				}
				err = server.Send(&service.HelloResponse{Msg: data})
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
