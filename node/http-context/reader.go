package http_context

import (
	"io"
	"strings"

	"github.com/eolinker/eosc/log"
)

type Reader struct {
	reader io.Reader

	agent     *requestAgent
	record    strings.Builder
	requestId string
	resp      *Response
}

func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err != nil {
		log.Debug("read error:", err)
		log.DebugF("request id %s ,read body: %s", r.requestId, r.record.String())
		return 0, err
	}
	r.record.Write(p[:n])
	if r.agent != nil {
		r.agent.responseBody.Write(p[:n])
	}
	if r.resp != nil {
		r.resp.AppendBody(p[:n])
	}
	return n, nil
}
