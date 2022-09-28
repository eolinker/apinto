package params_transformer

import (
	"encoding/json"
	"errors"
	"fmt"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"mime/multipart"
	"net/url"
	"strings"
)

const (
	FormParamType string = "application/x-www-form-urlencoded"
	JsonType      string = "application/json"
	MultipartType string = "multipart/form-data"

	paramConvert = "convert"
	paramError   = "error"
	paramOrigin  = "origin"

	servereErrStatusCode = 500
	clientErrStatusCode  = 400
	successStatusCode    = 200
)

var (
	paramPositionErrInfo  = `[plugin params-transformer config err] param position must be in the set ["query","header",body]. err position: %s `
	paramNameErrInfo      = `[plugin params-transformer config err] param name must be not null. `
	paramProxyNameErrInfo = `[plugin params-transformer config err] param proxy_name must be not null. `
)

func encodeErr(ent string, origin string, code int) error {
	if ent == "json" {
		tmp := map[string]interface{}{
			"message":     origin,
			"status_code": code,
		}
		info, _ := json.Marshal(tmp)
		return fmt.Errorf("%s", info)
	}
	return fmt.Errorf("%s statusCode: %d", origin, code)
}

func parseBodyParams(ctx http_service.IHttpContext, contentType string) (interface{}, map[string][]string, map[string][]*multipart.FileHeader, error) {

	if strings.Contains(contentType, FormParamType) {
		formParams, err := ctx.Proxy().Body().BodyForm()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("[params_transformer] parse body error: %w", err)
		}
		return nil, formParams, nil, nil
	} else if strings.Contains(contentType, JsonType) {
		body, err := ctx.Proxy().Body().RawBody()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("[params_transformer] get body error: %w", err)
		}
		if string(body) == "" {
			body = []byte("{}")
		}
		obj, err := oj.Parse(body)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("[params_transformer] parse body error: %w", err)
		}
		return obj, nil, nil, nil
	} else if strings.Contains(contentType, MultipartType) {
		formParams, err := ctx.Proxy().Body().BodyForm()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("[params_transformer] parse body error: %w", err)
		}
		files, err := ctx.Proxy().Body().Files()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("[params_transformer] parse body error: %w", err)
		}
		return nil, formParams, files, nil
	}

	return nil, nil, nil, errors.New("[params_transformer] unsupported content-type: " + contentType)
}

func getHeaderValue(paramName string, ctx http_service.IHttpContext, isRemove bool) (string, error) {
	paramValue := ctx.Proxy().Header().GetHeader(paramName)

	if len(paramValue) < 1 {
		errInfo := "[params_transformer] param " + paramName + " required"
		return "", errors.New(errInfo)
	}

	if isRemove {
		ctx.Proxy().Header().DelHeader(paramName)
	}

	return paramValue, nil
}

func getQueryValue(paramName string, ctx http_service.IHttpContext, isRemove bool) (string, error) {
	paramValue := ctx.Proxy().URI().GetQuery(paramName)

	if len(paramValue) < 1 {
		errInfo := "[params_transformer] param " + paramName + " required"
		return "", errors.New(errInfo)
	}

	if isRemove {
		ctx.Proxy().URI().DelQuery(paramName)
	}

	return paramValue, nil
}

func getBodyValue(bh *bodyHandler, paramName, contentType string, ctx http_service.IHttpContext, isRemove bool) (*bodyHandler, interface{}, error) {
	if bh == nil {
		bodyParams, formParams, files, err := parseBodyParams(ctx, contentType)
		if err != nil {
			return nil, nil, err
		}
		bh = &bodyHandler{
			body:       bodyParams,
			formParams: formParams,
			files:      files,
		}
	}
	var value interface{} = nil
	errInfo := "[params_transformer] param " + paramName + " required"

	if strings.Contains(contentType, FormParamType) {

		if v, ok := bh.formParams[paramName]; ok {
			if isRemove {
				delete(bh.formParams, paramName)
			}
			return bh, v, nil
		}
		return bh, nil, errors.New(errInfo)
	} else if strings.Contains(contentType, JsonType) {
		if !strings.HasPrefix(paramName, "$.") {
			paramName = fmt.Sprintf("$.%s", paramName)
		}
		x, err := jp.ParseString(paramName)
		if err != nil {
			return bh, nil, fmt.Errorf("[params_transformer] fail to get body,error: %w", err)
		}
		value := x.Get(bh.body)
		if isRemove {
			err = x.Del(bh.body)
			if err != nil {
				return bh, nil, fmt.Errorf("[params_transformer] fail to get body,error: %w", err)
			}
		}
		if len(value) > 0 {
			return bh, value[0], nil
		}
		return bh, "", nil
	} else if strings.Contains(contentType, MultipartType) {
		value, ok := bh.formParams[paramName]
		if !ok {
			fileValue, fileOk := bh.files[paramName]
			if !fileOk {
				return bh, "", errors.New(errInfo)
			}
			if isRemove {
				delete(bh.files, paramName)
			}
			return bh, fileValue, nil

		}
		if isRemove {
			delete(bh.formParams, paramName)
		}
		return bh, value, nil
	}
	return bh, value, nil
}

func getProxyValue(position, proxyPosition, contentType string, headerValue, queryValue string, bodyValue interface{}) (string, interface{}, error) {
	value := ""
	var bodyContent interface{} = nil
	if position == "header" {
		value = headerValue
	} else if position == "query" {
		value = queryValue
	} else if position == "body" {
		if strings.Contains(contentType, FormParamType) || strings.Contains(contentType, MultipartType) {
			if v, ok := bodyValue.([]string); ok {
				value = v[0]
			} else {
				bodyContent = bodyValue
			}
		} else if strings.Contains(contentType, JsonType) {
			if proxyPosition == "body" {
				bodyContent = bodyValue
			} else {
				v, _ := json.Marshal(bodyValue)
				value = string(v)
			}
		}
	}
	return value, bodyContent, nil
}

type bodyHandler struct {
	formParams url.Values
	body       interface{}
	files      map[string][]*multipart.FileHeader
}
