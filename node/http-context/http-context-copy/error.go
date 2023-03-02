package http_context_copy

import "errors"

var (
	ErrorNotForm      = errors.New("contentType is not Form")
	ErrorNotMultipart = errors.New("contentType is not Multipart")
	ErrorNotAllowRaw  = errors.New("contentType is not allow Raw")
	ErrorNotSend      = errors.New("not send")
)
