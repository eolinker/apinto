package main

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"encoding/json"
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/eolinker/apinto/utils"
	"reflect"
	"time"
)

func client(addr string, serviceName, methodName string, timeout time.Duration, typesList []string, valuesList []hessian.Object) (interface{}, error) {
	arguments := make([]interface{}, 3)
	parameterValues := make([]reflect.Value, 3)

	arguments[0] = methodName
	arguments[1] = typesList
	arguments[2] = valuesList

	parameterValues[0] = reflect.ValueOf(arguments[0])
	parameterValues[1] = reflect.ValueOf(arguments[1])
	parameterValues[2] = reflect.ValueOf(arguments[2])

	invoc := invocation.NewRPCInvocationWithOptions(invocation.WithMethodName("$invoke"),
		invocation.WithArguments(arguments),
		invocation.WithParameterValues(parameterValues))

	url, err := common.NewURL(addr,
		common.WithProtocol(dubbo.DUBBO), common.WithParamsValue(constant.SerializationKey, constant.Hessian2Serialization),
		//common.WithParamsValue(constant.GenericFilterKey, "true"),
		common.WithParamsValue(constant.TimeoutKey, timeout.String()),
		common.WithParamsValue(constant.InterfaceKey, serviceName),
		//common.WithParamsValue(constant.ReferenceFilterKey, "generic"),
		//dubboAttachment must contains group and version info
		common.WithPath(serviceName),
	)
	if err != nil {
		return nil, err
	}
	dubboProtocol := dubbo.NewDubboProtocol()

	invoker := dubboProtocol.Refer(url)
	var resp interface{}
	invoc.SetReply(&resp)

	result := invoker.Invoke(context.Background(), invoc)
	if result.Error() != nil {
		return nil, result.Error()
	}

	return result.Result(), nil
}

func main() {
	ComplexServer()
	//List()
	//GetById(101)
}

func ComplexServer() {
	var types []string
	var valuesList []hessian.Object

	types = append(types, "object")

	server := map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"}

	valuesList = append(valuesList, map[string]interface{}{"time": time.Now(), "addr": "192.168.0.1", "server": server})

	resp, err := client("172.18.166.219:20000", "api.Server", "ComplexServer", time.Second*3, types, valuesList)

	if err != nil {
		fmt.Println(err)
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)

	bytes, err := json.Marshal(vvv)

	fmt.Println(string(bytes), err)
}

func UpdateList() {
	var types []string
	var valuesList []hessian.Object

	types = append(types, "object")
	val1 := map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"}
	val2 := map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"}
	valuesList = append(valuesList, []interface{}{val1, val2})

	resp, err := client("127.0.0.1:20000", "api.Server", "UpdateList", time.Second*3, types, valuesList)

	if err != nil {
		fmt.Println(err)
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)

	bytes, err := json.Marshal(vvv)

	fmt.Println(string(bytes), err)
}

func Update() {
	var types []string
	var valuesList []hessian.Object

	types = append(types, "object")
	valuesList = append(valuesList, map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"})
	resp, err := client("127.0.0.1:20000", "api.Server", "Update", time.Second*3, types, valuesList)

	if err != nil {
		fmt.Println(err)
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)

	bytes, err := json.Marshal(vvv)

	fmt.Println(string(bytes), err)
}

func List() {
	var types []string
	var valuesList []hessian.Object

	types = append(types, "object")
	valuesList = append(valuesList, map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"})
	resp, err := client("127.0.0.1:20000", "api.Server", "List", time.Second*3, types, valuesList)

	if err != nil {
		fmt.Println(err)
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)

	bytes, err := json.Marshal(vvv)

	fmt.Println(string(bytes), err)
}

func GetById(id int64) {
	types := make([]string, 0)
	valuesList := make([]hessian.Object, 0)

	types = append(types, "int64")
	valuesList = append(valuesList, id)

	resp, err := client("127.0.0.1:20000", "api.Server", "GetById", time.Second*3, types, valuesList)

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
