package main

import (
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	http_dubbo "github.com/eolinker/apinto/app/dubbo-test/http-dubbo"
	"time"
)

func main() {
	//go dubbo_server.StartDubboServer()

	time.Sleep(time.Second)

	//http_dubbo.TcpToDubbo()
	//return
	types := make([]string, 0)
	types = append(types, "object")
	valuesList := make([]hessian.Object, 0)

	valuesList = append(valuesList, map[string]interface{}{"name": "张泽意啊啊啊"})
	dubbo, err := http_dubbo.HttpToDubbo("dubbo://127.0.0.1:20001", "api.UserService", "GetUser", types, valuesList)
	i := dubbo.(*interface{})
	fmt.Println()
	fmt.Println(*i, err)
}
