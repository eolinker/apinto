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
	respTypeErrInfo         = `[plugin extra-params config err] responseType must be in the set ["text","json"]. err responseType: %s `
	paramPositionErrInfo    = `[plugin extra-params config err] param position must be in the set ["query","header",body]. err position: %s `
	conflictSolutionErrInfo = `[plugin extra-params config err] param conflictSolution must be in the set ["origin","convert","error"]. err conflictSolution: %s`
	parseBodyErrInfo        = `[extra_params] Fail to parse body! [err]: %s`
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
	return fmt.Errorf("%s statusCode: %s", origin, statusCode)
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
	paramName := ConvertHearderKey(param.ParamName)
	if _, ok := param.ParamValue.(string); !ok {
		errInfo := "[extra_params] Header param " + param.ParamName + " must be a string"
		return "", errors.New(errInfo)
	}
	if param.ParamConflictSolution == "" {
		param.ParamConflictSolution = paramConvert
	}

	var paramValue string

	if _, ok := headers[paramName]; !ok {
		param.ParamConflictSolution = paramConvert
	} else {
		paramValue = headers[paramName][0]
	}

	if param.ParamConflictSolution == paramConvert {
		if value, ok := param.ParamValue.(string); ok {
			paramValue = value
		} else {
			errInfo := `[extra_params] Illegal "paramValue" in "` + param.ParamName + `"`
			return "", errors.New(errInfo)
		}
	} else if param.ParamConflictSolution == paramError {
		errInfo := `[extra_params] "` + param.ParamName + `" has a conflict.`
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
	if _, ok := param.ParamValue.(string); !ok {
		errInfo := "[extra_params] Query param " + param.ParamName + " must be a string"
		return "", errors.New(errInfo)
	}
	value := ""
	if param.ParamConflictSolution == "" {
		param.ParamConflictSolution = paramConvert
	}

	//判断请求中是否包含对应的query参数
	if !hasQueryValue(ctx.Proxy().URI().RawQuery(), param.ParamName) {
		param.ParamConflictSolution = paramConvert
	} else {
		value = ctx.Proxy().URI().GetQuery(param.ParamName)
	}

	if param.ParamConflictSolution == paramConvert {
		value = param.ParamValue.(string)
	} else if param.ParamConflictSolution == paramError {
		errInfo := `[extra_params] "` + param.ParamName + `" has a conflict.`
		return "", errors.New(errInfo)
	}

	return value, nil
}

func getBodyValue(bodyParams map[string]interface{}, formParams map[string][]string, param *ExtraParam, contentType string) (interface{}, error) {
	var value interface{} = nil
	if param.ParamConflictSolution == "" {
		param.ParamConflictSolution = paramConvert
	}
	if strings.Contains(contentType, FormParamType) {
		if _, ok := param.ParamValue.(string); !ok {
			errInfo := "[extra_params] Body param " + param.ParamName + " must be a string"
			return "", errors.New(errInfo)
		}
		if _, ok := formParams[param.ParamName]; !ok {
			param.ParamConflictSolution = paramConvert
		} else {
			value = formParams[param.ParamName][0]
		}
	} else if strings.Contains(contentType, JsonType) {
		if _, ok := bodyParams[param.ParamName]; !ok {
			param.ParamConflictSolution = paramConvert
		} else {
			value = bodyParams[param.ParamName]
		}
	}
	if param.ParamConflictSolution == paramConvert {
		value = param.ParamValue
	} else if param.ParamConflictSolution == paramError {
		errInfo := `[extra_params] "` + param.ParamName + `" has a conflict.`
		return "", errors.New(errInfo)
	}

	return value, nil
}

func ConvertHearderKey(header string) string {
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
