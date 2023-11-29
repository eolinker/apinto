package params_transformer

import (
	"encoding/json"
	"mime"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/ohler55/ojg/jp"

	"github.com/eolinker/apinto/drivers"
)

var _ http_service.HttpFilter = (*ParamsTransformer)(nil)
var _ eocontext.IFilter = (*ParamsTransformer)(nil)

type ParamsTransformer struct {
	drivers.WorkerBase
	params    []*TransParam
	remove    bool
	errorType string
}

func (p *ParamsTransformer) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(p, ctx, next)
}

func (p *ParamsTransformer) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {
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

	contentType, _, _ := mime.ParseMediaType(ctx.Proxy().Body().ContentType())
	var bh *bodyHandler = nil

	for _, param := range p.params {
		headerValue := ""
		queryValue := ""
		var bodyValue interface{} = nil
		switch param.Position {
		case "header":
			{
				var err error
				headerValue, err = getHeaderValue(param.Name, ctx, p.remove)
				if err != nil {
					if param.Required {
						err = encodeErr(p.errorType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
				}
			}
		case "query":
			{
				var err error
				queryValue, err = getQueryValue(param.Name, ctx, p.remove)
				if err != nil {
					if param.Required {
						err = encodeErr(p.errorType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
				}
			}
		case "body":
			{
				var err error
				bh, bodyValue, err = getBodyValue(bh, param.Name, contentType, ctx, p.remove)
				if err != nil {
					if param.Required {
						err = encodeErr(p.errorType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
				}
			}
		}

		switch param.ProxyPosition {
		case "header":
			{
				value, _, err := getProxyValue(param.Position, param.ProxyPosition, contentType, headerValue, queryValue, bodyValue)
				if err != nil {
					err = encodeErr(p.errorType, err.Error(), clientErrStatusCode)
					return clientErrStatusCode, err
				}
				if ctx.Proxy().Header().GetHeader(param.ProxyName) != "" {
					ctx.Proxy().Header().AddHeader(param.ProxyName, value)
				} else {
					ctx.Proxy().Header().SetHeader(param.ProxyName, value)
				}
			}
		case "query":
			{
				value, _, err := getProxyValue(param.Position, param.ProxyPosition, contentType, headerValue, queryValue, bodyValue)
				if err != nil {
					err = encodeErr(p.errorType, err.Error(), clientErrStatusCode)
					return clientErrStatusCode, err
				}

				if ctx.Proxy().URI().GetQuery(param.ProxyName) != "" {
					ctx.Proxy().URI().AddQuery(param.ProxyName, value)
				} else {
					ctx.Proxy().URI().SetQuery(param.ProxyName, value)
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

				value, bv, err := getProxyValue(param.Position, param.ProxyPosition, contentType, headerValue, queryValue, bodyValue)
				if err != nil {
					err = encodeErr(p.errorType, err.Error(), clientErrStatusCode)
					return clientErrStatusCode, err
				}
				if strings.Contains(contentType, FormParamType) {
					if _, ok := bh.formParams[param.ProxyName]; ok {
						bh.formParams[param.ProxyName] = append(bh.formParams[param.ProxyName], value)
					} else {
						bh.formParams[param.ProxyName] = []string{value}
					}
				} else if strings.Contains(contentType, JsonType) {
					paramName := param.ProxyName
					if !strings.HasPrefix(paramName, "$.") {
						paramName = "$." + paramName
					}

					x, err := jp.ParseString(paramName)
					if err != nil {
						err = encodeErr(p.errorType, err.Error(), clientErrStatusCode)
						return clientErrStatusCode, err
					}
					if param.Position != "body" {
						x.Set(bh.body, value)
						continue
					}
					x.Set(bh.body, bv)
				} else if strings.Contains(contentType, MultipartType) {
					if len(value) > 0 {
						if _, ok := bh.formParams[param.ProxyName]; ok {
							bh.formParams[param.ProxyName] = append(bh.formParams[param.ProxyName], value)
						} else {
							bh.formParams[param.ProxyName] = []string{value}
						}
					} else {
						//ctx.Proxy().AddFile(param.ProxyName, bv.(*apinto_plugin.FileHeader))
						bh.files[param.ProxyName] = bv.([]*multipart.FileHeader)
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

func (p *ParamsTransformer) Start() error {
	return nil
}

func (p *ParamsTransformer) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	confObj, err := check(conf)
	if err != nil {
		return err
	}

	p.params = confObj.Params
	p.remove = confObj.Remove
	p.errorType = confObj.ErrorType

	return nil
}

func (p *ParamsTransformer) Stop() error {
	return nil
}

func (p *ParamsTransformer) Destroy() {
	p.params = nil
	p.errorType = ""
}

func (p *ParamsTransformer) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}
