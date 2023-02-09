package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"liujian-test/grpc-test-demo/service"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"
	"google.golang.org/grpc" // 引入grpc认证包
)

const (
	// Address gRPC服务地址
	//Address = "120.25.14.89:9999"
	Address = "www.choosy-liu.cn:9001"
	//Address = "172.18.189.43:9001"
)

func CurrentRequest(names []string, md map[string]string, authority string) error {
	var err error
	var opts []grpc.DialOption

	//opts = append(opts, grpc.WithDefaultCallOptions())
	//opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))
	cerds := credentials.NewTLS(&tls.Config{})
	//cerds, err := credentials.NewClientTLSFromFile("../nginx.pem", "")
	//if err != nil {
	//	return err
	//}
	opts = append(opts, grpc.WithTransportCredentials(cerds))
	//opts = append(opts, grpc.WithAuthority(authority))
	//}
	// 指定客户端interceptor
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor))

	conn, err := grpc.Dial(Address, opts...)
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
		_, err = c.Hello(ctx, &service.HelloRequest{Name: name}, grpc.Header(&header), grpc.Trailer(&trailer))
		fmt.Println("err:", err)
		fmt.Println("header:", header)
		fmt.Println("trailing:", trailer)
	}

	return nil
}

func GetStreamClient(authority string) (*grpc.ClientConn, service.HelloClient, error) {
	var err error
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())
	//opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))
	opts = append(opts, grpc.WithAuthority(authority))
	//}

	// 指定客户端interceptor
	opts = append(opts, grpc.WithStreamInterceptor(streamInterceptor))

	conn, err := grpc.Dial(Address, opts...)
	if err != nil {
		return nil, nil, err
	}

	// 初始化客户端
	return conn, service.NewHelloClient(conn), nil
}

func StreamRequest(names []string, md map[string]string, authority string) error {
	conn, c, err := GetStreamClient(authority)
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

func StreamResponse(names []string, md map[string]string, authority string) error {
	conn, c, err := GetStreamClient(authority)
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

func AllStream(names []string, md map[string]string, authority string) error {
	conn, c, err := GetStreamClient(authority)
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
