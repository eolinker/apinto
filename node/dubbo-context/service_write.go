package dubbo_context

import (
	dubbo_context "github.com/eolinker/eosc/eocontext/dubbo-context"
	"time"
)

var _ dubbo_context.IServiceWriter = (*RequestServiceWrite)(nil)

type RequestServiceWrite struct {
	path        string
	serviceName string
	group       string
	version     string
	method      string
	timeout     time.Duration
}

func (r *RequestServiceWrite) Path() string {
	return r.path
}

func (r *RequestServiceWrite) Interface() string {
	return r.serviceName
}

func (r *RequestServiceWrite) Group() string {
	return r.group
}

func (r *RequestServiceWrite) Version() string {
	return r.version
}

func (r *RequestServiceWrite) Method() string {
	return r.method
}

func (r *RequestServiceWrite) Timeout() time.Duration {
	return r.timeout
}

func (r *RequestServiceWrite) SetPath(path string) {
	r.path = path

}

func (r *RequestServiceWrite) SetInterface(s string) {
	r.serviceName = s
}

func (r *RequestServiceWrite) SetGroup(group string) {
	r.group = group
}

func (r *RequestServiceWrite) SetVersion(s string) {
	r.version = s
}

func (r *RequestServiceWrite) SetMethod(s string) {
	r.method = s
}

func (r *RequestServiceWrite) SetTimeout(t time.Duration) {
	r.timeout = t
}
