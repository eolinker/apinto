package auth

import http_context "github.com/eolinker/goku-eosc/node/http-context"

type IAuth interface {
	Auth(ctx *http_context.Context) error
}
