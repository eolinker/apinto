package dubbo2_context

import (
	dubbo2_context "github.com/eolinker/eosc/eocontext/dubbo2-context"
)

var _ dubbo2_context.IRequestReader = (*RequestReader)(nil)

type RequestReader struct {
	serviceReader dubbo2_context.IServiceReader
	body          interface{}
	host          string
	remoteIp      string
	attachments   map[string]interface{}
}

func NewRequestReader(serviceReader dubbo2_context.IServiceReader, host string, remoteIp string, attachments map[string]interface{}) *RequestReader {
	return &RequestReader{serviceReader: serviceReader, host: host, remoteIp: remoteIp, attachments: attachments}
}

func (r *RequestReader) RemoteIP() string {
	return r.remoteIp
}

func (r *RequestReader) Host() string {
	return r.host
}

func (r *RequestReader) Attachments() map[string]interface{} {
	return r.attachments
}

func (r *RequestReader) Attachment(s string) (interface{}, bool) {
	v, ok := r.attachments[s]
	return v, ok
}

func (r *RequestReader) Service() dubbo2_context.IServiceReader {
	return r.serviceReader
}

func (r *RequestReader) Body() interface{} {
	return r.body
}
