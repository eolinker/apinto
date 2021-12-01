package extra_params

import (
	"encoding/json"
	"errors"
	"fmt"
	http_service "github.com/eolinker/eosc/http-service"
	"strings"
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

func parseBodyParams(ctx http_service.IHttpContext, body []byte, contentType string) (map[string]interface{}, map[string][]string, error) {
	formParams := make(map[string][]string)
	bodyParams := make(map[string]interface{})
	var err error
	if strings.Contains(contentType, FormParamType) {
		formParams, err = ctx.Proxy().Body().BodyForm()
		if err != nil {
			return bodyParams, formParams, err
		}
	} else if strings.Contains(contentType, JsonType) {
		if string(body) != "" {
			err = json.Unmarshal(body, &bodyParams)
			if err != nil {
				return bodyParams, formParams, err
			}
		}
	}

	return bodyParams, formParams, nil
}

func getHeaderValue(headers map[string][]string, param *ExtraParam) (string, error) {
	paramName := ConvertHeaderKey(param.Name)
	if _, ok := param.Value.(string); !ok {
		errInfo := "[extra_params] Header param " + param.Name + " must be a string"
		return "", errors.New(errInfo)
	}
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
		if value, ok := param.Value.(string); ok {
			paramValue = value
		} else {
			errInfo := `[extra_params] Illegal "paramValue" in "` + param.Name + `"`
			return "", errors.New(errInfo)
		}
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

func getQueryValue(ctx http_service.IHttpContext, param *ExtraParam) (string, error) {
	if _, ok := param.Value.(string); !ok {
		errInfo := "[extra_params] Query param " + param.Name + " must be a string"
		return "", errors.New(errInfo)
	}
	value := ""
	if param.Conflict == "" {
		param.Conflict = paramConvert
	}

	//判断请求中是否包含对应的query参数
	if !hasQueryValue(ctx.Proxy().URI().RawQuery(), param.Name) {
		param.Conflict = paramConvert
	} else {
		value = ctx.Proxy().URI().GetQuery(param.Name)
	}

	if param.Conflict == paramConvert {
		value = param.Value.(string)
	} else if param.Conflict == paramError {
		errInfo := `[extra_params] "` + param.Name + `" has a conflict.`
		return "", errors.New(errInfo)
	}

	return value, nil
}

func getBodyValue(bodyParams map[string]interface{}, formParams map[string][]string, param *ExtraParam, contentType string) (interface{}, error) {
	var value interface{} = nil
	if param.Conflict == "" {
		param.Conflict = paramConvert
	}
	if strings.Contains(contentType, FormParamType) {
		if _, ok := param.Value.(string); !ok {
			errInfo := "[extra_params] Body param " + param.Name + " must be a string"
			return "", errors.New(errInfo)
		}
		if _, ok := formParams[param.Name]; !ok {
			param.Conflict = paramConvert
		} else {
			value = formParams[param.Name][0]
		}
	} else if strings.Contains(contentType, JsonType) {
		if _, ok := bodyParams[param.Name]; !ok {
			param.Conflict = paramConvert
		} else {
			value = bodyParams[param.Name]
		}
	}
	if param.Conflict == paramConvert {
		value = param.Value
	} else if param.Conflict == paramError {
		errInfo := `[extra_params] "` + param.Name + `" has a conflict.`
		return "", errors.New(errInfo)
	}

	return value, nil
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
