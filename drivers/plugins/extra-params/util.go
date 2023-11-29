package extra_params

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/ohler55/ojg/oj"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	http_context "github.com/eolinker/apinto/node/http-context"
)

const (
	FormParamType string = "application/x-www-form-urlencoded"
	JsonType      string = "application/json"

	paramConvert string = "convert"
	paramError   string = "error"
	paramOrigin  string = "origin"

	serverErrStatusCode = 500
	clientErrStatusCode = 400
	successStatusCode   = 200
)

var (
	paramPositionErrInfo = `[plugin extra-params config err] param position must be in the set ["query","header",body]. err position: %s `
	parseBodyErrInfo     = `[extra_params] Fail to parse body! [err]: %s`
	paramNameErrInfo     = `[plugin params-transformer config err] param name must be not null. `
)

func encodeErr(ent string, origin string, statusCode int) error {
	if ent == "json" {
		tmp := map[string]interface{}{
			"message":     origin,
			"status_code": statusCode,
		}
		info, _ := json.Marshal(tmp)
		return fmt.Errorf("%s", info)
	}
	return fmt.Errorf("%s statusCode: %d", origin, statusCode)
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
	return nil, nil, errors.New("unsupported content-type: " + contentType)
}

//
//func parseBodyParams(ctx http_service.IHttpContext) (map[string]interface{}, map[string][]string, error) {
//	contentType, _, _ := mime.ParseMediaType(ctx.Proxy().Body().ContentType())
//
//	switch contentType {
//	case http_context.FormData, http_context.MultipartForm:
//		formParams, err := ctx.Proxy().Body().BodyForm()
//		if err != nil {
//			return nil, nil, err
//		}
//		return nil, formParams, nil
//	case http_context.JSON:
//		body, err := ctx.Proxy().Body().RawBody()
//		if err != nil {
//			return nil, nil, err
//		}
//		var bodyParams map[string]interface{}
//		err = json.Unmarshal(body, &bodyParams)
//		if err != nil {
//			return bodyParams, nil, err
//		}
//	}
//	return nil, nil, errors.New("[params_transformer] unsupported content-type: " + contentType)
//}

func getHeaderValue(headers map[string][]string, param *ExtraParam, value string) (string, error) {
	paramName := ConvertHeaderKey(param.Name)

	if param.Conflict == "" {
		param.Conflict = paramConvert
	}

	var paramValue string

	if _, ok := headers[paramName]; !ok {
		param.Conflict = paramConvert
	} else {
		paramValue = headers[paramName][0]
	}

	if param.Conflict == paramConvert {
		paramValue = value
	} else if param.Conflict == paramError {
		errInfo := `[extra_params] "` + param.Name + `" has a conflict.`
		return "", errors.New(errInfo)
	}

	return paramValue, nil
}

func hasQueryValue(rawQuery string, paramName string) bool {
	bytes := []byte(rawQuery)
	if len(bytes) == 0 {
		return false
	}

	k := 0
	for i, c := range bytes {
		switch c {
		case '=':
			key := string(bytes[k:i])
			if key == paramName {
				return true
			}
		case '&':
			k = i + 1
		}
	}

	return false
}

func getQueryValue(ctx http_service.IHttpContext, param *ExtraParam, value string) (string, error) {
	paramValue := ""
	if param.Conflict == "" {
		param.Conflict = paramConvert
	}

	//判断请求中是否包含对应的query参数
	if !hasQueryValue(ctx.Proxy().URI().RawQuery(), param.Name) {
		param.Conflict = paramConvert
	} else {
		paramValue = ctx.Proxy().URI().GetQuery(param.Name)
	}

	if param.Conflict == paramConvert {
		paramValue = value
	} else if param.Conflict == paramError {
		errInfo := `[extra_params] "` + param.Name + `" has a conflict.`
		return "", errors.New(errInfo)
	}

	return paramValue, nil
}

func getBodyValue(bodyParams map[string]interface{}, formParams map[string][]string, param *ExtraParam, contentType string, value interface{}) (interface{}, error) {
	var paramValue interface{} = nil
	Conflict := param.Conflict
	if Conflict == "" {
		Conflict = paramConvert
	}
	if strings.Contains(contentType, http_context.FormData) || strings.Contains(contentType, http_context.MultipartForm) {
		if _, ok := formParams[param.Name]; !ok {
			Conflict = paramConvert
		} else {
			paramValue = formParams[param.Name][0]
		}
	} else if strings.Contains(contentType, http_context.JSON) {
		if _, ok := bodyParams[param.Name]; !ok {
			param.Conflict = paramConvert
		} else {
			paramValue = bodyParams[param.Name]
		}
	}
	if Conflict == paramConvert {
		paramValue = value
	} else if Conflict == paramError {
		errInfo := `[extra_params] "` + param.Name + `" has a conflict.`
		return "", errors.New(errInfo)
	}

	return paramValue, nil
}

func ConvertHeaderKey(header string) string {
	header = strings.ToLower(header)
	headerArray := strings.Split(header, "-")
	h := ""
	arrLen := len(headerArray)
	for i, value := range headerArray {
		vLen := len(value)
		if vLen < 1 {
			continue
		} else {
			if vLen == 1 {
				h += strings.ToUpper(value)
			} else {
				h += strings.ToUpper(string(value[0])) + value[1:]
			}
			if i != arrLen-1 {
				h += "-"
			}
		}
	}
	return h
}
