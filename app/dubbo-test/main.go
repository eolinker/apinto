package main

import (
	"encoding/json"
	"errors"
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	dubbo_server "github.com/eolinker/apinto/app/dubbo-test/dubbo-server"
	http_dubbo "github.com/eolinker/apinto/app/dubbo-test/http-dubbo"
	"github.com/eolinker/apinto/utils"
	"time"
)

var errClientReadTimeout = errors.New("maybe the client read timeout or fail to decode tcp stream in Writer.Write")

func main() {
	go dubbo_server.StartDubboServer()

	time.Sleep(time.Second)

	//http_dubbo.TcpToDubbo()
	//return
	types := make([]string, 0)
	types = append(types, "object")
	valuesList := make([]hessian.Object, 0)

	valuesList = append(valuesList, map[string]interface{}{"name": "123456", "id": 10})
	//valuesList = append(valuesList, "zhangzeyi")
	//cn.zzy.
	resp, err := http_dubbo.ProxyToDubbo("192.168.198.165:8099", "api.UserService", "GetUser", time.Second*3, types, valuesList)
	if err != nil {
		fmt.Println(err)
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)

	bytes, err := json.Marshal(vvv)
	fmt.Println(string(bytes), err)
}

func formatData(value interface{}) interface{} {

	switch valueTemp := value.(type) {
	case map[interface{}]interface{}:
		maps := make(map[string]interface{})
		for k, v := range valueTemp {
			maps[utils.InterfaceToString(k)] = formatData(v)
		}
		return maps
	case []interface{}:
		values := make([]interface{}, 0)

		for _, v := range valueTemp {
			values = append(values, formatData(v))
		}
		return values
	default:
		return value
	}
}
