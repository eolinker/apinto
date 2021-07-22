package http_context

import "net/url"

//Request 转发内容
type Request struct {
	*Header
	*CookiesHandler
	*BodyRequestHandler
	queries      url.Values
	targetURL    string
	targetServer string
	Method       string
	Scheme       string
}

func (r *Request) Querys() url.Values {
	return r.queries
}

//TargetURL 获取转发url
func (r *Request) TargetURL() string {
	return r.targetURL
}

//SetTargetURL 设置转发URL
func (r *Request) SetTargetURL(targetURL string) {
	r.targetURL = targetURL
}

//TargetServer 获取转发服务器地址
func (r *Request) TargetServer() string {
	return r.targetServer
}

//SetTargetServer 设置最终转发地址
func (r *Request) SetTargetServer(targetServer string) {
	r.targetServer = targetServer
}

//Queries 获取query参数
func (r *Request) Queries() url.Values {
	return r.queries
}

//NewRequest 创建请求
func NewRequest(r *RequestReader) *Request {
	if r == nil {
		return nil
	}
	header := r.Headers()
	return &Request{
		Scheme:             r.Proto(),
		Method:             r.Method(),
		Header:             NewHeader(header),
		CookiesHandler:     newCookieHandle(header),
		BodyRequestHandler: r.BodyRequestHandler.Clone(),
		queries:            r.URL().Query(),
	}
}
