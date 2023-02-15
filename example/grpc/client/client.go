package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"time"

	service "github.com/eolinker/apinto/example/grpc/demo_service"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"
	"google.golang.org/grpc" // 引入grpc认证包
)

func genDialOptions() ([]grpc.DialOption, error) {
	var opts []grpc.DialOption
	if insecureVerify {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})))
	} else {
		if keyFile != "" && certFIle != "" {
			certs, err := credentials.NewClientTLSFromFile(certFIle, "")
			if err != nil {
				return nil, err
			}
			opts = append(opts, grpc.WithTransportCredentials(certs))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}
	}
	if authority != "" {
		opts = append(opts, grpc.WithAuthority(authority))
	}
	return opts, nil
}

func CurrentRequest(names []string, md map[string]string) error {
	opts, err := genDialOptions()
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 初始化客户端
	c := service.NewHelloClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))
	var header, trailer metadata.MD
	for _, name := range names {
		// 调用方法
		response, err := c.Hello(ctx, &service.HelloRequest{Name: name}, grpc.Header(&header), grpc.Trailer(&trailer))
		fmt.Println("err:", err)
		fmt.Println("header:", header)
		fmt.Println("trailing:", trailer)
		fmt.Println("msg:", response.GetMsg())
	}

	return nil
}

func GetStreamClient() (*grpc.ClientConn, service.HelloClient, error) {
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

func StreamRequest(names []string, md map[string]string) error {
	conn, c, err := GetStreamClient()
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))
	var header, trailer metadata.MD
	// 调用方法
	client, err := c.StreamRequest(ctx, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return err
	}
	defer func() {
		reply, err := client.CloseAndRecv()
		fmt.Println("err:", err)
		fmt.Println("header:", header)
		fmt.Println("trailing:", trailer)
		fmt.Println("reply", reply)
	}()
	for _, name := range names {
		err = client.Send(&service.HelloRequest{Name: name})
		if err != nil {
			fmt.Println(err)
		}
	}
	time.Sleep(5 * time.Second)
	return nil
}

func StreamResponse(names []string, md map[string]string) error {
	conn, c, err := GetStreamClient()
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))

	// 调用方法
	for _, name := range names {
		var header, trailer metadata.MD
		client, err := c.StreamResponse(ctx, &service.HelloRequest{Name: name}, grpc.Header(&header), grpc.Trailer(&trailer))
		if err != nil {
			return err
		}
		doAcceptResponse(client, header, trailer)
	}

	return nil
}

func AllStream(names []string, md map[string]string) error {
	conn, c, err := GetStreamClient()
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))

	var header, trailer metadata.MD
	client, err := c.AllStream(ctx, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return err
	}
	// 调用方法
	for _, name := range names {
		err = client.Send(&service.HelloRequest{Name: name})
		if err != nil {
			fmt.Println(err)
		}
	}
	doAcceptResponse(client, header, trailer)
	return nil
}

func doAcceptResponse(client service.Hello_StreamResponseClient, header, trailer metadata.MD) {
	data := make(map[string]string)
	defer func() {
		err := client.CloseSend()
		if err != nil {
			fmt.Println("close stream response error:", err)
		}
		fmt.Println("header:", header)
		fmt.Println("trailing:", trailer)
		fmt.Println("reply", data)

	}()
	for {
		reply, err := client.Recv()
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			return
		}
		now := time.Now()
		data[now.Format("2006-01-02 15:04:05.000")] = reply.Msg
	}
}
