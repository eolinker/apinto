package main

import "flag"

var (
	address = "127.0.0.1:8099"
)

func init() {
	flag.StringVar(&address, "addr", "172.28.187.118:8099", "The address to connect dubbo2 server.")
	flag.Parse()
}
