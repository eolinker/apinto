package dubbo2_context

import (
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
)

var _ dubbo2_context.IServiceReader = (*RequestServiceReader)(nil)

type RequestServiceReader struct {
	path        string
	serviceName string
	group       string
	version     string
	method      string
}

func NewRequestServiceReader(path string, serviceName string, group string, version string, method string) *RequestServiceReader {
	return &RequestServiceReader{path: path, serviceName: serviceName, group: group, version: version, method: method}
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
