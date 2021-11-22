package params_transformer

import (
	"encoding/json"
	"github.com/eolinker/eosc"
	http_service "github.com/eolinker/eosc/http-service"
	"github.com/ohler55/ojg/jp"
	"strconv"
	"strings"
)

var _ http_service.IFilter = (*ParamsTransformer)(nil)

type ParamsTransformer struct {
	*Driver
	id                     string
	name                   string
	params                 []*TransParam
	removeAfterTransformed bool
	responseType           string
}

func (p *ParamsTransformer) DoFilter(ctx http_service.IHttpContext, next http_service.IChain) error {
	statusCode, err := p.access(ctx)
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

func (p *ParamsTransformer) access(ctx http_service.IHttpContext) (int, error) {

	contentType := ctx.Proxy().Header().GetHeader("Content-Type")
	var bh *bodyHandler = nil

	for _, param := range p.params {
		headerValue := ""
		queryValue := ""
		var bodyValue interface{} = nil
		switch param.ParamPosition {
		case "header":
			{
				var err error
				headerValue, err = getHeaderValue(param.ParamName, ctx, p.removeAfterTransformed)
				if err != nil {
					if param.Required {
						err = encodeErr(p.responseType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
				}
			}
		case "query":
			{
				var err error
				queryValue, err = getQueryValue(param.ParamName, ctx, p.removeAfterTransformed)
				if err != nil {
					if param.Required {
						err = encodeErr(p.responseType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
				}
			}
		case "body":
			{
				var err error
				bh, bodyValue, err = getBodyValue(bh, param.ParamName, contentType, ctx, p.removeAfterTransformed)
				if err != nil {
					if param.Required {
						err = encodeErr(p.responseType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
				}
			}
		}

		switch param.ProxyParamPosition {
		case "header":
			{
				value, _, err := getProxyValue(param.ParamPosition, param.ProxyParamPosition, contentType, headerValue, queryValue, bodyValue)
				if err != nil {
					err = encodeErr(p.responseType, err.Error(), clientErrStatusCode)
					return clientErrStatusCode, err
				}
				if ctx.Proxy().Header().GetHeader(param.ProxyParamName) != "" {
					ctx.Proxy().Header().AddHeader(param.ProxyParamName, value)
				} else {
					ctx.Proxy().Header().SetHeader(param.ProxyParamName, value)
				}
			}
		case "query":
			{
				value, _, err := getProxyValue(param.ParamPosition, param.ProxyParamPosition, contentType, headerValue, queryValue, bodyValue)
				if err != nil {
					err = encodeErr(p.responseType, err.Error(), clientErrStatusCode)
					return clientErrStatusCode, err
				}

				if ctx.Proxy().URI().GetQuery(param.ProxyParamName) != "" {
					ctx.Proxy().URI().AddQuery(param.ProxyParamName, value)
				} else {
					ctx.Proxy().URI().SetQuery(param.ProxyParamName, value)
				}

			}
		case "body":
			{
				if bh == nil {
					bodyParams, formParams, files, err := parseBodyParams(ctx, contentType)
					if err != nil {
						return clientErrStatusCode, err
					}
					bh = &bodyHandler{
						body:       bodyParams,
						formParams: formParams,
						files:      files,
					}
				}

				value, bv, err := getProxyValue(param.ParamPosition, param.ProxyParamPosition, contentType, headerValue, queryValue, bodyValue)
				if err != nil {
					err = encodeErr(p.responseType, err.Error(), clientErrStatusCode)
					return clientErrStatusCode, err
				}
				if strings.Contains(contentType, FormParamType) {
					if _, ok := bh.formParams[param.ProxyParamName]; ok {
						bh.formParams[param.ProxyParamName] = append(bh.formParams[param.ProxyParamName], value)
					} else {
						bh.formParams[param.ProxyParamName] = []string{value}
					}
				} else if strings.Contains(contentType, JsonType) {
					paramName := param.ProxyParamName
					if !strings.HasPrefix(paramName, "$.") {
						paramName = "$." + paramName
					}

					x, err := jp.ParseString(paramName)
					if err != nil {
						err = encodeErr(p.responseType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
					x.Set(bh.body, bv)
				} else if strings.Contains(contentType, MultipartType) {
					if len(value) > 0 {
						if _, ok := bh.formParams[param.ProxyParamName]; ok {
							bh.formParams[param.ProxyParamName] = append(bh.formParams[param.ProxyParamName], value)
						} else {
							bh.formParams[param.ProxyParamName] = []string{value}
						}
					} else {
						//ctx.Proxy().AddFile(param.ProxyParamName, bv.(*goku_plugin.FileHeader))
						bh.files[param.ProxyParamName] = bv.(*http_service.FileHeader)
					}
				} else {
					continue
				}
			}
		}
	}

	if bh != nil {
		if strings.Contains(contentType, FormParamType) {
			ctx.Proxy().Body().SetForm(bh.formParams)

		} else if strings.Contains(contentType, JsonType) {
			bodyByte, _ := json.Marshal(bh.body)
			ctx.Proxy().Body().SetRaw(contentType, bodyByte)
		}
	}

	return successStatusCode, nil
}

func (p *ParamsTransformer) Id() string {
	return p.id
}

func (p *ParamsTransformer) Start() error {
	return nil
}

func (p *ParamsTransformer) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	confObj, err := p.check(conf)
	if err != nil {
		return err
	}

	p.params = confObj.Params
	p.removeAfterTransformed = confObj.RemoveAfterTransformed
	p.responseType = confObj.ResponseType

	return nil
}

func (p *ParamsTransformer) Stop() error {
	return nil
}

func (p *ParamsTransformer) Destroy() {
	p.params = nil
	p.responseType = ""
}

func (p *ParamsTransformer) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
