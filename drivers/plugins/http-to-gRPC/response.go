package http_to_grpc

func NewResponse() *Response {
	return &Response{}
}

type Response struct {
	buf []byte
}

func (r *Response) Write(p []byte) (n int, err error) {
	r.buf = p
	return len(r.buf), nil
}

func (r *Response) Body() []byte {
	return r.buf
}
