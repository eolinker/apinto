package extra_params

import (
	"encoding/json"
	"fmt"
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"strconv"
	"strings"
)

var _ http_service.IFilter = (*ExtraParams)(nil)

type ExtraParams struct {
	*Driver
	id           string
	name         string
	params       []*ExtraParam
	responseType string
}

func (e *ExtraParams) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) error {
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
	contentType := ctx.Proxy().Header().GetHeader("Content-Type")

	body, _ := ctx.Proxy().Body().RawBody()
	bodyParams, formParams, err := parseBodyParams(ctx, body, contentType)
	if err != nil {
		errinfo := fmt.Sprintf(parseBodyErrInfo, err.Error())
		err = encodeErr(e.responseType, errinfo, serverErrStatusCode)
		return serverErrStatusCode, err
	}

	headers := ctx.Proxy().Header().Headers()
	// 先判断参数类型
	for _, param := range e.params {
		switch param.ParamPosition {
		case "query":
			{
				value, err := getQueryValue(ctx, param)
				if err != nil {
					err = encodeErr(e.responseType, err.Error(), serverErrStatusCode)
					return serverErrStatusCode, err
				}
				ctx.Proxy().URI().SetQuery(param.ParamName, value)
			}
		case "header":
			{
				value, err := getHeaderValue(headers, param)
				if err != nil {
					err = encodeErr(e.responseType, err.Error(), serverErrStatusCode)
					return serverErrStatusCode, err
				}
				ctx.Proxy().Header().SetHeader(param.ParamName, value)
			}
		case "body":
			{
				value, err := getBodyValue(bodyParams, formParams, param, contentType)
				if err != nil {
					err = encodeErr(e.responseType, err.Error(), serverErrStatusCode)
					return serverErrStatusCode, err
				}
				if strings.Contains(contentType, FormParamType) {
					err = ctx.Proxy().Body().SetToForm(param.ParamName, value.(string))
					if err != nil {
						err = encodeErr(e.responseType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
				} else if strings.Contains(contentType, JsonType) {
					bodyParams[param.ParamName] = value
				}
			}
		}
	}
	if strings.Contains(contentType, JsonType) {
		b, _ := json.Marshal(bodyParams)
		ctx.Proxy().Body().SetRaw(contentType, b)
	}

	return successStatusCode, nil
}

func (e *ExtraParams) Id() string {
	return e.id
}

func (e *ExtraParams) Start() error {
	return nil
}

func (e *ExtraParams) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	confObj, err := e.check(conf)
	if err != nil {
		return err
	}

	e.params = confObj.Params
	e.responseType = confObj.ResponseType

	return nil
}

func (e *ExtraParams) Stop() error {
	return nil
}

func (e *ExtraParams) Destroy() {
	e.params = nil
	e.responseType = ""
}

func (e *ExtraParams) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
