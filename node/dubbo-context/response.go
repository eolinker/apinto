package dubbo_context

import (
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
	"time"
)

var _ dubbo_context.IResponse = (*Response)(nil)

type Response struct {
	responseError error
	timeout       time.Duration
	duration      time.Duration
	body          []byte
}

func (r *Response) ResponseError() error {
	return r.responseError
}

func (r *Response) SetResponseTime(duration time.Duration) {
	r.duration = duration
}

func (r *Response) ResponseTime() time.Duration {
	return r.duration
}

func (r *Response) GetBody() []byte {
	return r.body
}

func (r *Response) BodyLen() int {
	return len(r.body)
}

func (r *Response) SetBody(body []byte) {
	r.body = body
}
