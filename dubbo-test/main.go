package main

import (
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	dubbo_server "github.com/eolinker/apinto/dubbo-test/dubbo-server"
	http_dubbo "github.com/eolinker/apinto/dubbo-test/http-dubbo"
	"time"
)

func main() {
	go dubbo_server.StartDubboServer()
	time.Sleep(time.Second)
	types := make([]string, 0)
	types = append(types, "java.lang.String")
	valuesList := make([]hessian.Object, 0)
	m := make(map[int]interface{})
	m[0] = "zhangzeyi"
	valuesList = append(valuesList, m)
	dubbo, err := http_dubbo.HttpToDubbo("dubbo://127.0.0.1:4399", "cn.zzy.api.UserService", "sayHello", types, valuesList)
	fmt.Println(&dubbo, err)
}
