package http_to_dubbo2

import (
	"encoding/json"
	"errors"
	"fmt"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"time"
)

var (
	ErrorTimeoutComplete = errors.New("complete timeout")
)

type Complete struct {
	retry   int
	timeOut time.Duration
	service string
	method  string
	params  []param
}

func NewComplete(retry int, timeOut time.Duration, service string, method string, params []param) *Complete {
	return &Complete{retry: retry, timeOut: timeOut, service: service, method: method, params: params}
}

func (c *Complete) Complete(org eocontext.EoContext) error {
	ctx, err := http_service.Assert(org)
	if err != nil {
		return err
	}
	//设置响应开始时间
	proxyTime := time.Now()

	balance := ctx.GetBalance()
	var lastErr error

	defer func() {
		if lastErr != nil {
			ctx.Response().SetStatus(400, "400")
			ctx.Response().SetBody([]byte(lastErr.Error()))
		}
		ctx.Response().SetResponseTime(time.Now().Sub(proxyTime))
		ctx.SetLabel("handler", "proxy")
	}()
	body, _ := ctx.Request().Body().RawBody()

	var types []string
	var valuesList []hessian.Object

	for _, v := range c.params {
		types = append(types, v.className)
	}

	//从body中提取内容
	if len(c.params) == 1 && c.params[0].fieldName == "" {
		var val interface{}

		if lastErr = json.Unmarshal(body, &val); lastErr != nil {
			log.Errorf("doHttpFilter jsonUnmarshal err:%v body:%v", lastErr, body)
			return lastErr
		}

		valuesList = append(valuesList, val)
	} else if len(c.params) == 1 && c.params[0].fieldName != "" {
		var maps map[string]interface{}

		if lastErr = json.Unmarshal(body, &maps); lastErr != nil {
			log.Errorf("doHttpFilter jsonUnmarshal err:%v body:%v", lastErr, body)
			return lastErr
		}

		if val, ok := maps[c.params[0].fieldName]; ok {
			valuesList = append(valuesList, val)
		} else {
			lastErr = errors.New(fmt.Sprintf("参数解析错误，body中未包含%s的参数名", c.params[0].fieldName))
			return lastErr
		}

	} else {
		var maps map[string]interface{}

		if lastErr = json.Unmarshal(body, &maps); lastErr != nil {
			log.Errorf("doHttpFilter jsonUnmarshal err:%v body:%v", lastErr, body)
			return lastErr
		}

		for _, v := range c.params {
			if val, ok := maps[v.fieldName]; ok {
				valuesList = append(valuesList, val)
			} else {
				lastErr = errors.New(fmt.Sprintf("参数解析错误，body中未包含%s的参数名", c.params[0].fieldName))
				return lastErr
			}
		}

	}

	client := newDubbo2Client(c.service, c.method, types, valuesList)

	for index := 0; index <= c.retry; index++ {

		if c.timeOut > 0 && time.Now().Sub(proxyTime) > c.timeOut {
			return ErrorTimeoutComplete
		}
		node, err := balance.Select(ctx)
		if err != nil {
			log.Error("select error: ", err)
			ctx.Response().SetStatus(501, "501")
			ctx.Response().SetBody([]byte(err.Error()))
			return err
		}

		log.Debug("node: ", node.Addr())
		var result interface{}
		result, lastErr = client.dial(ctx.Context(), node.Addr(), c.timeOut)
		if lastErr == nil {
			bytes, err := json.Marshal(result)
			if err != nil {
				lastErr = err
				return err
			}
			ctx.Response().SetBody(bytes)
			return nil
		}
		log.Error("http to dubbo2 dial error: ", lastErr)
	}

	return lastErr
}
