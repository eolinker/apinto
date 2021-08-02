package main

import "flag"

var (
	httpPort   = 8081
	httpsPort  = 8082
	path       = ""
	driverFile = "profession.yml"
)

func initFlag() {
	flag.IntVar(&httpPort, "http", 8081, "Please provide a valid http port")
	flag.IntVar(&httpsPort, "https", 8082, "Please provide a valid https port")

	flag.StringVar(&path, "path", "", "Please provide a valid file path")
	flag.StringVar(&path, "driver_path", "profession.yml", "Please provide a valid file path")

	flag.Parse()
}
