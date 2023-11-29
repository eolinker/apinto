package dubbo2_to_http

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo/impl"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/eolinker/eosc/eocontext"
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/utils"
)

var (
	errorTimeoutComplete = errors.New("complete timeout")

	errParamLen   = errors.New("args.length != types.length")
	errNodeIsNull = errors.New("node is null")
)

type Complete struct {
	retry       int
	timeOut     time.Duration
	contentType string
	path        string
	method      string
	params      []param
}

func NewComplete(retry int, timeOut time.Duration, contentType string, path string, method string, params []param) *Complete {
	return &Complete{retry: retry, timeOut: timeOut, contentType: contentType, path: path, method: method, params: params}
}

func (c *Complete) Complete(org eocontext.EoContext) error {
	ctx, err := dubbo2_context.Assert(org)
	if err != nil {
		return err
	}

	paramBody := ctx.Proxy().GetParam()
	if len(paramBody.TypesList) != len(paramBody.ValuesList) || len(c.params) != len(paramBody.TypesList) {
		ctx.Response().SetBody(Dubbo2ErrorResult(errParamLen))
		return err
	}

	paramMap := make(map[string]interface{})
	for i := range paramBody.ValuesList {
		paramMap[paramBody.TypesList[i]] = paramBody.ValuesList[i]
	}

	log.DebugF("dubbo2-to-http complete paramMap = %v", paramMap)

	//设置响应开始时间
	proxyTime := time.Now()
	defer func() {
		ctx.Response().SetResponseTime(time.Since(proxyTime))
	}()

	var reqBody []byte

	if len(paramMap) == 1 && c.params[0].fieldName != "" {
		object, ok := paramMap[c.params[0].className]
		if !ok {
			err = fmt.Errorf("参数解析错误，未找到的名称为 %s className", c.params[0].className)
			ctx.Response().SetBody(Dubbo2ErrorResult(err))
			return err
		}

		object = formatData(object)

		maps := make(map[string]hessian.Object)
		maps[c.params[0].fieldName] = object

		bytes, err := json.Marshal(maps)
		if err != nil {
			ctx.Response().SetBody(Dubbo2ErrorResult(err))
			return err
		}
		reqBody = bytes
	} else if len(paramMap) == 1 && c.params[0].fieldName == "" {
		object, ok := paramMap[c.params[0].className]
		if !ok {
			err = fmt.Errorf("参数解析错误，未找到的名称为 %s className", c.params[0].className)
			ctx.Response().SetBody(Dubbo2ErrorResult(err))
			return err
		}

		log.DebugF("dubbo2-to-http complete paramMap = %v params[0] = %v object=%v", paramMap, c.params[0], object)

		object = formatData(object)

		bytes, err := json.Marshal(object)
		if err != nil {
			log.Errorf("dubbo2-to-http complete err=%v", err)
			ctx.Response().SetBody(Dubbo2ErrorResult(err))
			return err
		}
		reqBody = bytes
	} else {
		maps := make(map[string]hessian.Object)
		for _, p := range c.params {
			object, ok := paramMap[p.className]
			if !ok {
				err = fmt.Errorf("参数解析错误，未找到的名称为 %s className", p.className)
				ctx.Response().SetBody(Dubbo2ErrorResult(err))
				return err
			}
			object = formatData(object)

			maps[p.fieldName] = object
		}

		bytes, err := json.Marshal(maps)
		if err != nil {
			ctx.Response().SetBody(Dubbo2ErrorResult(err))
			return err
		}
		reqBody = bytes

	}

	balance := ctx.GetBalance()
	scheme := balance.Scheme()

	switch strings.ToLower(scheme) {
	case "", "tcp":
		scheme = "http"
	case "tsl", "ssl", "https":
		scheme = "https"

	}

	httpClient := NewClient(c.method, reqBody, c.path)

	timeOut := balance.TimeOut()

	var lastErr error
	for index := 0; index <= c.retry; index++ {

		if c.timeOut > 0 && time.Since(proxyTime) > c.timeOut {
			ctx.Response().SetBody(Dubbo2ErrorResult(errorTimeoutComplete))
			return errorTimeoutComplete
		}

		node, _, err := balance.Select(ctx)
		if err != nil {
			log.Error("select error: ", err)
			ctx.Response().SetBody(Dubbo2ErrorResult(errNodeIsNull))
			return err
		}

		var resBody []byte
		resBody, lastErr = send(httpClient, scheme, node, timeOut)
		if lastErr == nil {
			var val interface{}
			if err = json.Unmarshal(resBody, &val); err != nil {
				ctx.Response().SetBody(Dubbo2ErrorResult(err))
				return err
			}

			ctx.Response().SetBody(getResponse(val, ctx.Proxy().Attachments()))
			return nil
		}
		log.Error("dubbo upstream send error: ", lastErr)
	}

	ctx.Response().SetBody(Dubbo2ErrorResult(lastErr))

	return lastErr
}
func send(client *Client, scheme string, node eocontext.INode, timeOut time.Duration) ([]byte, error) {

	addr := fmt.Sprintf("%s://%s", scheme, node.Addr())
	log.Debug("node: ", addr)
	resBody, err := client.dial(addr, timeOut)
	if err != nil {
		node.Down()
		return nil, err
	}
	return resBody, err

}
func Dubbo2ErrorResult(err error) protocol.RPCResult {
	payload := impl.NewResponsePayload(nil, err, nil)
	return protocol.RPCResult{
		Attrs: payload.Attachments,
		Err:   payload.Exception,
		Rest:  payload.RspObj,
	}
}

func getResponse(obj interface{}, attachments map[string]interface{}) protocol.RPCResult {
	payload := impl.NewResponsePayload(obj, nil, attachments)
	return protocol.RPCResult{
		Attrs: payload.Attachments,
		Err:   payload.Exception,
		Rest:  payload.RspObj,
	}
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
