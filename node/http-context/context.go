package http_context

import (
	"encoding/json"
	"net/http"

	"github.com/go-basic/uuid"

	"github.com/valyala/fasthttp"

	goku_plugin "github.com/eolinker/goku-standard-plugin"

	access_field "github.com/eolinker/goku-eosc/node/common/access-field"
	"github.com/eolinker/goku-eosc/utils"
)

var _ goku_plugin.ContextProxy = (*Context)(nil)

//Context context
type Context struct {
	responseWriter *fasthttp.RequestCtx
	*CookiesHandler
	*PriorityHeader
	*StatusHandler
	*StoreHandler
	RequestOrg           *RequestReader
	ProxyRequest         *Request
	ProxyResponseHandler *ResponseReader
	Body                 []byte

	requestID string

	RestfulParam map[string]string
	LogFields    *access_field.Fields

	labels map[string]string
}

func (ctx *Context) Labels() map[string]string {
	if ctx.labels == nil {
		ctx.labels = map[string]string{}
	}
	return ctx.labels
}

func (ctx *Context) SetLabels(labels map[string]string) {
	if ctx.labels == nil {
		ctx.labels = make(map[string]string)
	}
	if labels != nil {
		for k, v := range labels {
			ctx.labels[k] = v
		}
	}
}

//Finish finish
func (ctx *Context) Finish() (statusCode int) {
	defer ctx.responseWriter.Done()
	header := ctx.PriorityHeader.header

	statusCode = ctx.StatusHandler.code
	if statusCode == 0 {
		statusCode = 504
	}
	ctx.LogFields.StatusCode = statusCode
	bodyAllowed := true
	switch {
	case statusCode >= 100 && statusCode <= 199:
		bodyAllowed = false
		break
	case statusCode == 204:
		bodyAllowed = false
		break
	case statusCode == 304:
		bodyAllowed = false
		break
	}

	if ctx.PriorityHeader.appendHeader != nil {
		for k, vs := range ctx.PriorityHeader.appendHeader.header {
			for _, v := range vs {
				header.Add(k, v)
			}
		}
	}

	if ctx.PriorityHeader.setHeader != nil {
		for k, vs := range ctx.PriorityHeader.setHeader.header {
			header.Del(k)
			for _, v := range vs {
				header.Add(k, v)
			}
		}
	}

	for k, vs := range ctx.PriorityHeader.header {
		if k == "Content-Length" {
			continue
			//vs = []string{strconv.Itoa(len(string(ctx.body)))}
		}
		for _, v := range vs {
			ctx.responseWriter.Response.Header.Add(k, v)
		}
	}
	if ctx.ProxyResponseHandler.header.Get("Content-Type") == "" {
		ctx.responseWriter.Response.Header.Set("Content-Type", "application/json")
	}

	if ctx.ProxyResponseHandler.header.Get("Content-Encoding") == "gzip" {
		body, err := utils.GzipCompress(ctx.Body)
		if err == nil {
			ctx.Body = body
		}
	}
	ctx.responseWriter.SetStatusCode(statusCode)
	ctx.LogFields.ResponseHeader = utils.HeaderToString(ctx.header)

	if !bodyAllowed {
		return statusCode
	}
	ctx.responseWriter.SetBody(ctx.Body)
	ctx.LogFields.ResponseMsg = string(ctx.Body)
	ctx.LogFields.ResponseMsgSize = len(ctx.Body)
	return statusCode
}

//RequestId 请求ID
func (ctx *Context) RequestId() string {
	return ctx.requestID
}

//NewContext 创建Context
func NewContext(requestCtx *fasthttp.RequestCtx) *Context {
	requestID := uuid.New()
	requestReader := NewRequestReader(requestCtx.Request)
	ctx := &Context{
		responseWriter:       requestCtx,
		CookiesHandler:       newCookieHandle(requestReader.header),
		PriorityHeader:       NewPriorityHeader(),
		StatusHandler:        NewStatusHandler(),
		StoreHandler:         NewStoreHandler(),
		RequestOrg:           requestReader,
		ProxyRequest:         NewRequest(requestReader),
		ProxyResponseHandler: nil,
		requestID:            requestID,
		LogFields:            access_field.NewFields(),
	}
	ctx.LogFields.RequestHeader = utils.HeaderToString(requestReader.Headers())
	ctx.LogFields.RequestMsg = string(ctx.RequestOrg.rawBody)
	ctx.LogFields.RequestMsgSize = len(ctx.RequestOrg.rawBody)
	ctx.LogFields.RequestUri = requestReader.req.URL.RequestURI()
	ctx.LogFields.RequestID = requestID
	return ctx
}

//SetProxyResponse 设置转发响应
func (ctx *Context) SetProxyResponse(response *http.Response) {

	ctx.SetProxyResponseHandler(newResponseReader(response))

}

//SetProxyResponseHandler 设置转发响应处理器
func (ctx *Context) SetProxyResponseHandler(response *ResponseReader) {
	ctx.ProxyResponseHandler = response
	if ctx.ProxyResponseHandler != nil {
		ctx.Body = ctx.ProxyResponseHandler.body
		ctx.SetStatus(ctx.ProxyResponseHandler.StatusCode(), ctx.ProxyResponseHandler.Status())
		ctx.header = ctx.ProxyResponseHandler.header
	}
}

//func (ctx *Context) Write(w http.ResponseWriter) {
//	if ctx.StatusCode() == 0 {
//		ctx.SetStatus(200, "200 ok")
//	}
//	if ctx.Body != nil {
//		w.Write(ctx.Body)
//	}
//
//	w.WriteHeader(ctx.StatusCode())
//
//}

//GetBody 获取请求body
func (ctx *Context) GetBody() []byte {
	return ctx.Body
}

//SetBody 设置body
func (ctx *Context) SetBody(data []byte) {
	ctx.Body = data
}

func (ctx *Context) SetError(err error) {
	result := map[string]string{
		"status": "error",
		"msg":    err.Error(),
	}
	errByte, _ := json.Marshal(result)
	ctx.Body = errByte
}

//ProxyResponse 返回响应
func (ctx *Context) ProxyResponse() goku_plugin.ResponseReader {
	return ctx.ProxyResponseHandler
}

//Request 获取原始请求
func (ctx *Context) Request() goku_plugin.RequestReader {
	return ctx.RequestOrg
}

//Proxy 代理
func (ctx *Context) Proxy() goku_plugin.Request {
	return ctx.ProxyRequest
}

func NotFound(ctx *Context) {
	ctx.responseWriter.SetStatusCode(404)
	ctx.responseWriter.SetBody([]byte("404 Not Found"))
}
