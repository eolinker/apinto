package main

import "flag"

var (
	address = "127.0.0.1:20000"
)

func init() {
	flag.StringVar(&address, "addr", "127.0.0.1:20000", "The address to connect dubbo2 server.")
	flag.Parse()
}
