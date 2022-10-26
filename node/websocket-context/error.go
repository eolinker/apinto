package websocket_context

import "errors"

var (
	ErrorNotForm      = errors.New("contentType is not Form")
	ErrorNotMultipart = errors.New("contentType is not Multipart")
	ErrorNotAllowRaw  = errors.New("contentType is not allow Raw")
	ErrorNotSend      = errors.New("not send")
)
