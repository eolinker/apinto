package output

import http_service "github.com/eolinker/eosc/http-service"

type IOutput interface {
	Output(http_service.IHttpContext) error
}
