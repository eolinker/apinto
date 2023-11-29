package extra_params_v2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/eolinker/eosc"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	dynamic_params "github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params"
)

type Config struct {
	Params          []*ExtraParam `json:"params" label:"参数列表"`
	RequestBodyType string        `json:"request_body_type" enum:"form-data,json,multipart-formdata" label:"请求体类型"`
	ErrorType       string        `json:"error_type" enum:"text,json" label:"报错输出格式"`
}

func (c *Config) doCheck() error {
	c.ErrorType = strings.ToLower(c.ErrorType)
	if c.ErrorType != "text" && c.ErrorType != "json" {
		c.ErrorType = "text"
	}

	for _, param := range c.Params {
		if param.Name == "" {
			return fmt.Errorf(paramNameErrInfo)
		}

		param.Position = strings.ToLower(param.Position)
		if param.Position != "query" && param.Position != "header" && param.Position != "body" {
			return fmt.Errorf(paramPositionErrInfo, param.Position)
		}

		param.Conflict = strings.ToLower(param.Conflict)
		if param.Conflict != paramOrigin && param.Conflict != paramConvert && param.Conflict != paramError {
			param.Conflict = paramConvert
		}
	}
	c.RequestBodyType = strings.ToLower(c.RequestBodyType)
	if contentTypeMap[c.RequestBodyType] == "" && c.RequestBodyType != "" {
		return fmt.Errorf("error body type: %s", c.RequestBodyType)
	}
	return nil
}

type ExtraParam struct {
	Name     string   `json:"name" label:"参数名"`
	Type     string   `json:"type" label:"参数类型" enum:"string,int,float,bool,$datetime,$md5,$timestamp,$concat,$hmac-sha256"`
	Position string   `json:"position" enum:"header,query,body" label:"参数位置"`
	Value    []string `json:"value" label:"参数值列表"`
	Conflict string   `json:"conflict" label:"参数冲突时的处理方式" enum:"origin,convert,error"`
}

type baseParam struct {
	header []*paramInfo
	query  []*paramInfo
	body   []*paramInfo
}

func generateBaseParam(params []*ExtraParam) *baseParam {
	b := &baseParam{
		header: make([]*paramInfo, 0),
		query:  make([]*paramInfo, 0),
		body:   make([]*paramInfo, 0),
	}
	for _, param := range params {
		switch param.Position {
		case positionHeader:
			b.header = append(b.header, newParamInfo(param.Name, param.Value, param.Type, param.Conflict))
		case positionQuery:
			b.query = append(b.query, newParamInfo(param.Name, param.Value, param.Type, param.Conflict))
		case positionBody:
			b.body = append(b.body, newParamInfo(param.Name, param.Value, param.Type, param.Conflict))
		}
	}
	return b
}

func newParamInfo(name string, value []string, typ string, conflict string) *paramInfo {
	d := &paramInfo{name: name, value: strings.Join(value, ""), conflict: conflict, valueType: typ}
	valueLen := len(d.value)
	if strings.HasPrefix(typ, "$") {
		factory, has := dynamic_params.Get(typ)
		if has {
			driver, err := factory.Create(name, value)
			if err == nil {
				d.driver = driver
			}
		}
	} else if valueLen > 1 && d.value[0] == '$' {
		// 系统变量
		d.systemValue = true
		d.value = d.value[1:valueLen]
	}
	return d
}

type paramInfo struct {
	name        string
	valueType   string
	systemValue bool
	value       string
	driver      dynamic_params.IDynamicDriver
	conflict    string
}

func (b *paramInfo) Build(ctx http_service.IHttpContext, contentType string, params interface{}) (string, error) {
	return b.build(ctx, contentType, params)
}

func (b *paramInfo) Parse(value string) (interface{}, error) {
	switch b.valueType {
	case "int":
		v, err := strconv.Atoi(value)
		return v, err
	case "float":
		v, err := strconv.ParseFloat(value, 64)
		return v, err
	case "bool":
		v, err := strconv.ParseBool(value)
		return v, err
	default:
		return value, nil
	}
}

func (b *paramInfo) build(ctx http_service.IHttpContext, contentType string, params interface{}) (string, error) {
	if b.driver == nil {
		if b.systemValue {
			return eosc.ReadStringFromEntry(ctx.GetEntry(), b.value), nil
		}
		return b.value, nil
	}
	value, err := b.driver.Generate(ctx, contentType, params)
	if err != nil {
		return "", err
	}
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int32, int64:
		return fmt.Sprintf("%d", v), nil
	}
	return "", nil
}
