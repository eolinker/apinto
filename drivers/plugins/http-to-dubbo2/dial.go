package http_to_dubbo2

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc/eocontext"
	"reflect"
	"time"
)

type dubbo2Client struct {
	serviceName string
	methodName  string
	typesList   []string
	valuesList  []hessian.Object
}

func newDubbo2Client(serviceName string, methodName string, typesList []string, valuesList []hessian.Object) *dubbo2Client {
	return &dubbo2Client{serviceName: serviceName, methodName: methodName, typesList: typesList, valuesList: valuesList}
}

func (d *dubbo2Client) dial(ctx context.Context, node eocontext.INode, timeout time.Duration) (interface{}, error) {
	arguments := make([]interface{}, 3)
	parameterValues := make([]reflect.Value, 3)

	arguments[0] = d.methodName
	arguments[1] = d.typesList
	arguments[2] = d.valuesList

	parameterValues[0] = reflect.ValueOf(arguments[0])
	parameterValues[1] = reflect.ValueOf(arguments[1])
	parameterValues[2] = reflect.ValueOf(arguments[2])

	invoc := invocation.NewRPCInvocationWithOptions(invocation.WithMethodName("$invoke"),
		invocation.WithArguments(arguments),
		invocation.WithParameterValues(parameterValues))

	serviceName := d.serviceName
	url, err := common.NewURL(node.Addr(),
		common.WithProtocol(dubbo.DUBBO), common.WithParamsValue(constant.SerializationKey, constant.Hessian2Serialization),
		common.WithParamsValue(constant.GenericFilterKey, "true"),
		common.WithParamsValue(constant.TimeoutKey, timeout.String()),
		common.WithParamsValue(constant.InterfaceKey, serviceName),
		common.WithParamsValue(constant.ReferenceFilterKey, "generic,filter"),
		common.WithPath(serviceName),
	)
	if err != nil {
		node.Down()
		return nil, err
	}

	dubboProtocol := dubbo.NewDubboProtocol()
	invoker := dubboProtocol.Refer(url)
	var resp interface{}
	invoc.SetReply(&resp)

	result := invoker.Invoke(ctx, invoc)
	if result.Error() != nil {
		return nil, result.Error()
	}

	val := result.Result().(*interface{})

	data := formatData(*val)

	return data, nil
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
