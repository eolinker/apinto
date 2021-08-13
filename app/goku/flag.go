package main

import "flag"

var (
	httpPort     = 8081
	httpsPort    = 8082
	httpsPemPath = ""
	httpsKeyPath = ""
	dataPath     = ""
	driverFile   = "profession.yml"
)

func initFlag() {
	flag.IntVar(&httpPort, "http", 8081, "Please provide a valid http port")
	flag.IntVar(&httpsPort, "https", 8082, "Please provide a valid https port")

	flag.StringVar(&httpsPemPath, "pem", "", "Please provide a valid pem file path")
	flag.StringVar(&httpsKeyPath, "key", "", "Please provide a valid key file path")

	flag.StringVar(&dataPath, "data_path", "", "Please provide a valid file path")
	flag.StringVar(&driverFile, "driver_path", "profession.yml", "Please provide a valid file path")

	flag.Parse()
}
