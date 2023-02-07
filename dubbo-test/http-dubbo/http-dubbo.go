package http_dubbo

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"github.com/apache/dubbo-go-hessian2"
	"reflect"
)

// HttpToDubbo
// addr:dubbo://192.168.198.160:20000
func HttpToDubbo(addr string, serviceName, methodName string, typesList []string, valuesList []hessian.Object) (interface{}, error) {
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
		common.WithProtocol(dubbo.DUBBO), common.WithParamsValue(constant.SerializationKey, constant.ProtobufSerialization),
		common.WithParamsValue(constant.GenericFilterKey, "true"),
		common.WithParamsValue(constant.TimeoutKey, "5s"),
		common.WithParamsValue(constant.InterfaceKey, serviceName),
		common.WithParamsValue(constant.ReferenceFilterKey, "generic,filter"),
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
