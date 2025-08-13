package main

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/logger"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"encoding/json"
	hessian "github.com/apache/dubbo-go-hessian2"
	"reflect"
	"strconv"
	"time"
)

func init() {
	logger.InitLogger(nil)
}

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
	//UpdateList()
	//Update()
}

func ComplexServer() {
	var types []string
	var valuesList []hessian.Object
	
	types = append(types, "object")
	
	server := map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"}
	
	valuesList = append(valuesList, map[string]interface{}{"time": time.Now(), "addr": "192.168.0.1", "server": server})
	resp, err := client(address, "api.Server", "ComplexServer", time.Second*3, types, valuesList)
	
	if err != nil {
		logger.Errorf("ComplexServer err=%s", err.Error())
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)
	
	bytes, _ := json.Marshal(vvv)
	logger.Infof("ComplexServer result=%s", string(bytes))
}

func UpdateList() {
	var types []string
	var valuesList []hessian.Object
	
	types = append(types, "object")
	val1 := map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"}
	val2 := map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"}
	valuesList = append(valuesList, []interface{}{val1, val2})
	
	resp, err := client(address, "api.Server", "UpdateList", time.Second*3, types, valuesList)
	
	if err != nil {
		logger.Errorf("UpdateList err=%s", err.Error())
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)
	
	bytes, _ := json.Marshal(vvv)
	logger.Infof("UpdateList result=%s", string(bytes))
}

func Update() {
	var types []string
	var valuesList []hessian.Object
	
	types = append(types, "object")
	valuesList = append(valuesList, map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"})
	resp, err := client(address, "api.Server", "Update", time.Second*3, types, valuesList)
	
	if err != nil {
		logger.Errorf("Update err=%s", err.Error())
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)
	
	bytes, _ := json.Marshal(vvv)
	
	logger.Infof("Update result=%s", string(bytes))
}

func List() {
	var types []string
	var valuesList []hessian.Object
	
	types = append(types, "object")
	valuesList = append(valuesList, map[string]interface{}{"id": 16, "name": "apinto", "email": "1324204490@qq.com"})
	resp, err := client(address, "api.Server", "List", time.Second*3, types, valuesList)
	
	if err != nil {
		logger.Errorf("List err=%s", err.Error())
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)
	
	bytes, _ := json.Marshal(vvv)
	
	logger.Infof("List result=%s", string(bytes))
}

func GetById(id int64) {
	types := make([]string, 0)
	valuesList := make([]hessian.Object, 0)
	
	types = append(types, "int64")
	valuesList = append(valuesList, id)
	
	resp, err := client(address, "api.Server", "GetById", time.Second*3, types, valuesList)
	
	if err != nil {
		logger.Errorf("List err=%s", err.Error())
		return
	}
	v := resp.(*interface{})
	vvv := formatData(*v)
	
	bytes, _ := json.Marshal(vvv)
	
	logger.Infof("GetById result=%s", string(bytes))
}

func formatData(value interface{}) interface{} {
	
	switch valueTemp := value.(type) {
	case map[interface{}]interface{}:
		maps := make(map[string]interface{})
		for k, v := range valueTemp {
			maps[InterfaceToString(k)] = formatData(v)
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

func InterfaceToString(value interface{}) string {
	var key string
	if value == nil {
		return key
	}
	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		return ""
	}
	return key
}
