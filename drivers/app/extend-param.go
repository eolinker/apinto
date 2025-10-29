package app

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/ohler55/ojg/oj"

	"github.com/ohler55/ojg/jp"

	http_context "github.com/eolinker/apinto/node/http-context"

	"github.com/eolinker/apinto/application"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type additionalParam struct {
	params []*Additional
}

func newAdditionalParam(params []*Additional) *additionalParam {
	return &additionalParam{params: params}
}

func (a *additionalParam) Execute(ctx http_service.IHttpContext) error {
	if len(a.params) < 1 {
		return nil
	}
	contentType, _, _ := mime.ParseMediaType(ctx.Proxy().Body().ContentType())
	var bodyParams interface{}
	var formParams map[string][]string
	var err error

	for _, p := range a.params {
		conflict := p.Conflict
		if conflict == "" {
			conflict = "convert"
		}
		switch p.Position {
		case application.PositionBody:
			if ctx.Proxy().Method() != http.MethodPost && ctx.Proxy().Method() != http.MethodPut && ctx.Proxy().Method() != http.MethodPatch {
				continue
			}
			if bodyParams == nil && formParams == nil {
				bodyParams, formParams, err = parseBodyParams(ctx)
				if err != nil {
					return fmt.Errorf(`fail to parse body! [err]: %v`, err)
				}
			}
			switch contentType {
			case http_context.FormData, http_context.MultipartForm:
				switch p.Conflict {
				case conflictConvert:
					formParams[p.Key] = []string{p.Value}
				case conflictOrigin, conflictError:
					{
						if _, ok := formParams[p.Key]; ok {
							if p.Conflict == conflictError {
								return fmt.Errorf(errorExist, p.Position, p.Key)
							}
						}
						formParams[p.Key] = []string{p.Value}
					}
				}
			case http_context.JSON:
				{
					key := p.Key
					if !strings.HasPrefix(p.Key, "$.") {
						key = "$." + key
					}
					x, err := jp.ParseString(key)
					if err != nil {
						return fmt.Errorf("parse key error: %v", err)
					}
					switch conflict {
					case conflictConvert:
						err = x.Set(bodyParams, p.Value)
						if err != nil {
							return fmt.Errorf("set additional json param error: %v", err)
						}
					case conflictOrigin, conflictError:
						{
							result := x.Get(bodyParams)
							if len(result) < 1 {
								err = x.Set(bodyParams, p.Value)
								if err != nil {
									return fmt.Errorf("set additional json param error: %v", err)
								}
							}
							if conflict == conflictError {
								return fmt.Errorf(errorExist, p.Position, p.Key)
							}
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

	if strings.Contains(contentType, http_context.FormData) || strings.Contains(contentType, http_context.MultipartForm) {
		ctx.Proxy().Body().SetForm(formParams)
	} else if strings.Contains(contentType, http_context.JSON) {
		b, _ := json.Marshal(bodyParams)
		ctx.Proxy().Body().SetRaw(contentType, b)
	}
	return nil
}

func parseBodyParams(ctx http_service.IHttpContext) (interface{}, map[string][]string, error) {
	if ctx.Proxy().Method() != http.MethodPost && ctx.Proxy().Method() != http.MethodPut && ctx.Proxy().Method() != http.MethodPatch {
		return nil, nil, nil
	}
	contentType, _, _ := mime.ParseMediaType(ctx.Proxy().Body().ContentType())
	switch contentType {
	case http_context.FormData, http_context.MultipartForm:
		formParams, err := ctx.Proxy().Body().BodyForm()
		if err != nil {
			return nil, nil, err
		}
		return nil, formParams, nil
	case http_context.JSON:
		body, err := ctx.Proxy().Body().RawBody()
		if err != nil {
			return nil, nil, err
		}
		if string(body) == "" {
			body = []byte("{}")
		}
		bodyParams, err := oj.Parse(body)
		return bodyParams, nil, err
	}
	return nil, nil, nil
}
