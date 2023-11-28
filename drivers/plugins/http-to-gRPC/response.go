package http_to_grpc

import (
	"strings"

	"go.uber.org/zap/buffer"
)

var (
	responseHeaderPre  = "\nResponse headers received:\n"
	responseContentPre = "\nResponse contents:\n"
	responseTrailerPre = "\nResponse trailers received:\n"
)

func NewResponse() *Response {
	return &Response{
		header:    make(map[string]string),
		bodyWrite: false,
		body:      &buffer.Buffer{},
	}
}

type Response struct {
	header    map[string]string
	bodyWrite bool
	body      *buffer.Buffer
}

func (r *Response) Write(p []byte) (n int, err error) {
	str := string(p)
	if strings.HasPrefix(str, responseHeaderPre) || strings.HasPrefix(str, responseTrailerPre) {
		str = strings.Replace(str, responseHeaderPre, "", 1)
		str = strings.Replace(str, responseTrailerPre, "", 1)
		headers := strings.Split(str, "\n")
		if len(headers) == 2 && strings.HasPrefix(headers[1], "(empty)") {
			return len(p), nil
		}
		for _, header := range headers {
			if strings.TrimSpace(header) == "" {
				continue
			}

			values := strings.Split(header, ":")
			var v string
			if len(values) > 1 {
				v = values[1]
				r.header[values[0]] = v
			}
		}
	}
	if strings.HasPrefix(str, responseContentPre) {
		r.bodyWrite = true
		return len(p), nil
	}
	if r.bodyWrite {
		r.body.Write(p)
		r.bodyWrite = false
	}
	return len(p), nil
}

func (r *Response) Body() []byte {
	return r.body.Bytes()
}

func (r *Response) Header() map[string]string {
	return r.header
}
