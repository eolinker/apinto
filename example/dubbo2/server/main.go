package main

import (
	_ "dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/imports"
)

func main() {
	config.SetProviderService(&Server{})
	err := config.Load(config.WithPath("./dubbogo.yaml"))
	if err != nil {
		panic(err)
	}
	select {}
}
