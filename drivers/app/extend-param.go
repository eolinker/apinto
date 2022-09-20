package app

import (
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	http_context "github.com/eolinker/apinto/node/http-context"

	"github.com/eolinker/apinto/application"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	FormParamType = "application/x-www-form-urlencoded"
	JsonType      = "application/json"
)

type additionalParam struct {
	params      []*Additional
	needRawBody bool
}

func newAdditionalParam(params []*Additional) *additionalParam {
	needRawBody := false
	for _, p := range params {
		if p.Position == application.PositionBody {
			needRawBody = true
			break
		}
	}
	return &additionalParam{params: params, needRawBody: needRawBody}
}

func (a *additionalParam) getBody(ctx http_service.IHttpContext) (string, bool) {
	needBody := false
	body := ""
	if a.needRawBody {
		method := ctx.Proxy().Method()
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
			b, _ := ctx.Proxy().Body().RawBody()

			needBody = true
			body = string(b)
		}
	}
	return body, needBody
}

func (a *additionalParam) Execute(ctx http_service.IHttpContext) error {
	body, needBody := a.getBody(ctx)
	for _, p := range a.params {
		conflict := p.Conflict
		if conflict == "" {
			conflict = "convert"
		}
		switch p.Position {
		case application.PositionBody:
			if !needBody {
				continue
			}
			contentType := ctx.Proxy().Header().GetHeader("Content-Type")
			if strings.Contains(contentType, http_context.FormData) || strings.Contains(contentType, http_context.MultipartForm) {
				switch p.Conflict {
				case conflictConvert:
					ctx.Proxy().Body().SetToForm(p.Key, p.Value)
				case conflictOrigin, conflictError:
					{
						v := ctx.Proxy().Body().GetForm(p.Key)
						if v == "" {
							ctx.Proxy().Body().SetToForm(p.Key, p.Value)
						} else {
							if p.Conflict == conflictError {
								return fmt.Errorf(errorExist, p.Position, p.Key)
							}
						}
					}
				}
			} else if strings.Contains(contentType, http_context.JSON) {
				obj, err := oj.ParseString(body)
				if err != nil {
					return fmt.Errorf("parse body error: %v", err)
				}
				key := p.Key
				if !strings.HasPrefix(p.Key, "$.") {
					key = "$." + key
				}
				x, err := jp.ParseString(p.Key)
				if err != nil {
					return fmt.Errorf("parse key error: %v", err)
				}
				switch conflict {
				case conflictConvert:
					err = x.Set(obj, p.Value)
					if err != nil {
						return fmt.Errorf("set additional json param error: %v", err)
					}
					body = x.String()
				case conflictOrigin, conflictError:
					{
						result := x.Get(p.Key)
						if len(result) < 1 {
							err = x.Set(obj, p.Value)
							if err != nil {
								return fmt.Errorf("set additional json param error: %v", err)
							}
							body = x.String()
						}
						if conflict == conflictError {
							return fmt.Errorf(errorExist, p.Position, p.Key)
						}
					}
				}
			}
		case application.PositionHeader:
			switch conflict {
			case conflictConvert:
				ctx.Proxy().Header().SetHeader(p.Key, p.Value)
			case conflictOrigin, conflictError:
				{
					_, has := ctx.Proxy().Header().Headers()[textproto.CanonicalMIMEHeaderKey(p.Key)]
					if !has {
						ctx.Proxy().Header().SetHeader(p.Key, p.Value)
					} else {
						if conflict == conflictError {
							return fmt.Errorf(errorExist, p.Position, p.Key)
						}
					}
				}
			}
		case application.PositionQuery:
			switch conflict {
			case conflictConvert:
				ctx.Proxy().URI().SetQuery(p.Key, p.Value)
			case conflictOrigin, conflictError:
				v := ctx.Proxy().URI().GetQuery(p.Key)
				if v == "" {
					ctx.Proxy().URI().SetQuery(p.Key, p.Value)
				} else {
					if conflict == conflictError {
						return fmt.Errorf(errorExist, p.Position, p.Key)
					}
				}
			}
		}
	}
	return nil
}
