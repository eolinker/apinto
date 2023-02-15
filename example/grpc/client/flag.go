package main

import "flag"

var (
	address        = "127.0.0.1:8099"
	authority      = ""
	keyFile        = ""
	certFIle       = ""
	serverName     = ""
	insecureVerify = false
)

func Parse() {
	flag.StringVar(&address, "addr", "127.0.0.1:8099", "The address to connect grpc server.")
	flag.StringVar(&authority, "authority", "", "Authority will be verified by grpc server.")
	flag.StringVar(&serverName, "-servername", "", "Override server name when validating TLS certificate.")
	flag.BoolVar(&insecureVerify, "insecure", false, "Skip server certificate and domain verification. (NOT SECURE!).")
	flag.StringVar(&keyFile, "key", "", "File containing client private key, to present to the server.")
	flag.StringVar(&certFIle, "cert", "", "File containing client certificate (public key).")
	flag.Parse()
}
