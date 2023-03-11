package main

import (
	"io"
	"log"
	"strings"
	"time"

	service "github.com/eolinker/apinto/example/grpc/demo_service"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"
	"google.golang.org/grpc" // 引入grpc认证包
)

type Client struct {
	client service.HelloClient
	conn   *grpc.ClientConn
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) CurrentRequest(names []string, md map[string]string) error {
	log.Println("start current request client,please wait...")
	defer log.Println("end current request")

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))
	var header, trailer metadata.MD
	for _, name := range names {
		// 调用方法
		response, err := c.client.Hello(ctx, &service.HelloRequest{Name: name}, grpc.Header(&header), grpc.Trailer(&trailer))
		log.Println("err:", err)
		log.Println("header:", header)
		log.Println("trailing:", trailer)
		log.Println("msg:", response.GetMsg())
	}

	return nil
}

func NewClient() (*grpc.ClientConn, service.HelloClient, error) {
	opts, err := genDialOptions()
	if err != nil {
		return nil, nil, err
	}
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, nil, err
	}

	// 初始化客户端
	return conn, service.NewHelloClient(conn), nil
}

func (c *Client) StreamRequest(names []string, md map[string]string) error {
	log.Println("start stream request client,please wait...")
	defer log.Println("end stream request")

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))
	var header, trailer metadata.MD
	// 调用方法
	client, err := c.client.StreamRequest(ctx, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return err
	}

	for _, name := range names {
		err = client.Send(&service.HelloRequest{Name: name})
		if err != nil {
			log.Println(err)
		}
	}
	reply, err := client.CloseAndRecv()
	log.Println("err:", err)
	log.Println("header:", header)
	log.Println("trailing:", trailer)
	log.Println("reply", reply.GetMsg())
	return nil
}

func (c *Client) StreamResponse(names []string, md map[string]string) error {
	log.Println("start stream response client,please wait...")
	defer log.Println("end stream response")
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))

	// 调用方法
	var header, trailer metadata.MD
	client, err := c.client.StreamResponse(ctx, &service.HelloRequest{Name: strings.Join(names, ",")}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return err
	}
	data := make(map[string]string)
	for {
		reply, err := client.Recv()
		if err != nil {
			if err != io.EOF {
				log.Println("err:", err)
			}
			break
		}
		now := time.Now()
		data[now.Format("2006-01-02 15:04:05.000")] = reply.Msg
	}

	log.Println("header:", header)
	log.Println("trailing:", trailer)
	log.Println("reply", data)
	return nil
}

func (c *Client) AllStream(names []string, md map[string]string) error {
	log.Println("start all stream client,please wait...")
	defer log.Println("end all stream")
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))

	var header, trailer metadata.MD
	client, err := c.client.AllStream(ctx, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return err
	}
	go func() {
		for {
			// 调用方法
			for _, name := range names {
				err = client.Send(&service.HelloRequest{Name: name})
				if err != nil {
					if err != io.EOF {
						log.Println("err:", err)
					}
					return
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	data := make(map[string]string)
	for {
		reply, err := client.Recv()
		if err != nil {
			if err != io.EOF {
				log.Println("err:", err)
			}
			break
		}
		now := time.Now()
		data[now.Format("2006-01-02 15:04:05.000")] = reply.Msg
	}

	log.Println("header:", header)
	log.Println("trailing:", trailer)
	log.Println("reply", data)
	return nil
}
