package dubbo2_context

import (
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
	"time"
)

var _ dubbo2_context.IServiceReader = (*RequestServiceReader)(nil)

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
