package extra_params

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/ohler55/ojg/jp"

	http_context "github.com/eolinker/apinto/node/http-context"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ http_service.HttpFilter = (*ExtraParams)(nil)
var _ eocontext.IFilter = (*ExtraParams)(nil)

var (
	errorExist = "%s: %s is already exists"
)

type ExtraParams struct {
	drivers.WorkerBase
	params    []*ExtraParam
	errorType string
}

func (e *ExtraParams) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *ExtraParams) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
	statusCode, err := e.access(ctx)
	if err != nil {
		ctx.Response().SetBody([]byte(err.Error()))
		ctx.Response().SetStatus(statusCode, strconv.Itoa(statusCode))
		return err
	}

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (e *ExtraParams) access(ctx http_service.IHttpContext) (int, error) {
	// 判断请求携带的content-type
	contentType, _, _ := mime.ParseMediaType(ctx.Proxy().Body().ContentType())

	bodyParams, formParams, err := parseBodyParams(ctx)
	if err != nil {
		errInfo := fmt.Sprintf(parseBodyErrInfo, err.Error())
		err = encodeErr(e.errorType, errInfo, serverErrStatusCode)
		return serverErrStatusCode, err
	}

	headers := ctx.Proxy().Header().Headers()
	// 先判断参数类型
	for _, param := range e.params {
		var paramValue interface{}
		err = json.Unmarshal([]byte(param.Value), &paramValue)
		if err != nil {
			paramValue = param.Value
		}
		switch param.Position {
		case "query":
			{
				v, _ := json.Marshal(paramValue)
				value, err := getQueryValue(ctx, param, string(v))
				if err != nil {
					err = encodeErr(e.errorType, err.Error(), clientErrStatusCode)
					return clientErrStatusCode, err
				}
				ctx.Proxy().URI().SetQuery(param.Name, value)
			}
		case "header":
			{
				v, _ := json.Marshal(paramValue)
				value, err := getHeaderValue(headers, param, string(v))
				if err != nil {
					err = encodeErr(e.errorType, err.Error(), clientErrStatusCode)
					return clientErrStatusCode, err
				}
				ctx.Proxy().Header().SetHeader(param.Name, value)
			}
		case "body":
			{
				if ctx.Proxy().Method() != http.MethodPost && ctx.Proxy().Method() != http.MethodPut && ctx.Proxy().Method() != http.MethodPatch {
					continue
				}
				switch contentType {
				case http_context.FormData, http_context.MultipartForm:
					if _, has := formParams[param.Name]; has {
						switch param.Conflict {
						case paramConvert:
							formParams[param.Name] = []string{paramValue.(string)}
						case paramOrigin:
						case paramError:
							return clientErrStatusCode, errors.New(`[extra_params] "` + param.Name + `" has a conflict.`)
						default:
							formParams[param.Name] = []string{paramValue.(string)}
						}
					} else {
						formParams[param.Name] = []string{paramValue.(string)}
					}
				case http_context.JSON:
					{
						key := param.Name
						if !strings.HasPrefix(param.Name, "$.") {
							key = "$." + key
						}
						x, err := jp.ParseString(key)
						if err != nil {
							return 400, fmt.Errorf("parse key error: %v", err)
						}
						switch param.Conflict {
						case paramConvert:
							err = x.Set(bodyParams, param.Value)
							if err != nil {
								return 400, fmt.Errorf("set additional json param error: %v", err)
							}
						case paramOrigin, paramError:
							{
								result := x.Get(bodyParams)
								if len(result) < 1 {
									err = x.Set(bodyParams, param.Value)
									if err != nil {
										return 400, fmt.Errorf("set additional json param error: %v", err)
									}
								}
								if param.Conflict == paramError {
									return 400, fmt.Errorf(errorExist, param.Position, param.Name)
								}
							}
						}
					}
				}
			}
			//if strings.Contains(contentType, http_context.FormData) || strings.Contains(contentType, http_context.MultipartForm) {
			//
			//} else if strings.Contains(contentType, ) {
			//	if _, has := bodyParams[param.Name]; has {
			//		switch param.Conflict {
			//		case paramConvert:
			//			bodyParams[param.Name] = paramValue.(string)
			//		case paramOrigin:
			//		case paramError:
			//			return clientErrStatusCode, errors.New(`[extra_params] "` + param.Name + `" has a conflict.`)
			//		default:
			//			bodyParams[param.Name] = paramValue
			//		}
			//	} else {
			//		bodyParams[param.Name] = paramValue
			//	}
			//}

		}
	}
	if strings.Contains(contentType, http_context.FormData) || strings.Contains(contentType, http_context.MultipartForm) {
		ctx.Proxy().Body().SetForm(formParams)
	} else if strings.Contains(contentType, http_context.JSON) {
		b, _ := json.Marshal(bodyParams)
		ctx.Proxy().Body().SetRaw(contentType, b)
	}

	return successStatusCode, nil
}

func (e *ExtraParams) Start() error {
	return nil
}

func (e *ExtraParams) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	confObj, err := check(conf)
	if err != nil {
		return err
	}

	e.params = confObj.Params
	e.errorType = confObj.ErrorType

	return nil
}

func (e *ExtraParams) Stop() error {
	return nil
}

func (e *ExtraParams) Destroy() {
	e.params = nil
	e.errorType = ""
}

func (e *ExtraParams) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
