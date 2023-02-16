package main

import (
	"crypto/tls"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"
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
