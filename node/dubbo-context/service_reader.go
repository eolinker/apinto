package dubbo_context

import (
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
	"time"
)

var _ dubbo_context.IServiceReader = (*RequestServiceReader)(nil)

type RequestServiceReader struct {
	path        string
	serviceName string
	group       string
	version     string
	method      string
	timeout     time.Duration
}

func (r *RequestServiceReader) Path() string {
	return r.path
}

func (r *RequestServiceReader) Interface() string {
	return r.serviceName
}

func (r *RequestServiceReader) Group() string {
	return r.group
}

func (r *RequestServiceReader) Version() string {
	return r.version
}

func (r *RequestServiceReader) Method() string {
	return r.method
}

func (r *RequestServiceReader) Timeout() time.Duration {
	return r.timeout
}
