package main

import "log"

func main() {
	Parse()
	conn, client, err := NewClient()
	if err != nil {
		log.Println("create grpc client error:", err)
		return
	}
	grpcClient := &Client{
		client: client,
		conn:   conn,
	}
	defer grpcClient.Close()
	md := map[string]string{
		"app": "apinto",
	}
	names := []string{
		"apinto",
		"eolink",
	}
	grpcClient.CurrentRequest(names, md)
	grpcClient.StreamRequest(names, md)
	grpcClient.StreamResponse(names, md)
	grpcClient.AllStream(names, md)
}
