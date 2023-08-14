package md5

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/ohler55/ojg/oj"

	"github.com/ohler55/ojg/jp"

	"github.com/eolinker/apinto/utils"

	"github.com/eolinker/eosc/log"
)

const (
	positionCurrent = iota
	positionHeader
	positionQuery
	// body
	positionBody
	positionSystem
)

type MD5 struct {
	name  string
	value []*Value
}

type Value struct {
	key      string
	position int
	optional bool
}

func NewMD5(name string, value []string) *MD5 {
	vs := make([]*Value, 0, len(value))
	for _, v := range value {
		v = strings.TrimSpace(v)
		vLen := len(v)
		if vLen > 0 {
			if v[0] == '{' && v[vLen-1] == '}' {
				vars := strings.Split(v[1:vLen-1], ".")
				position := positionBody
				variable := vars[0]
				if len(vars) > 1 {
					variable = vars[1]
					switch vars[0] {
					case "header":
						position = positionHeader
					case "query":
						position = positionQuery
					}
				}
				vs = append(vs, &Value{
					key:      variable,
					position: position,
				})
			} else if v[0] == '#' {
				vars := strings.Split(v[1:], ".")
				position := positionBody
				variable := vars[0]
				if len(vars) > 1 {
					variable = vars[1]
					switch vars[0] {
					case "header":
						position = positionHeader
					case "query":
						position = positionQuery
					}
				}
				vs = append(vs, &Value{
					key:      variable,
					position: position,
					optional: true,
				})
			} else if vLen > 3 && v[0] == '$' && v[1] == '{' && v[vLen-1] == '}' {
				// 使用系统变量
				vs = append(vs, &Value{
					key:      v[2 : vLen-1],
					position: positionSystem,
				})
			} else {
				vs = append(vs, &Value{
					key: v,
				})
			}
		}
	}
	return &MD5{
		name:  name,
		value: vs,
	}
}

func (m *MD5) Name() string {
	return m.name
}

func (m *MD5) Generate(ctx http_service.IHttpContext, contentType string, args ...interface{}) (interface{}, error) {
	builder := strings.Builder{}
	var params interface{}
	if contentType == "application/json" {
		if len(args) < 1 {
			return nil, errors.New("missing args")
		}
		params = args[0]
	}
	for _, v := range m.value {
		if v.key == "" {
			continue
		}
		builder.WriteString(retrieveParam(ctx, contentType, params, v))
	}
	log.Debug("md5 result: ", builder.String())
	if strings.HasPrefix(m.name, "__") {
		return utils.Md5(builder.String()), nil
	}
	return strings.ToUpper(utils.Md5(builder.String())), nil
}

func retrieveParam(ctx http_service.IHttpContext, contentType string, body interface{}, value *Value) string {
	switch value.position {
	case positionCurrent:
		return value.key
	case positionHeader:
		return ctx.Proxy().Header().Headers().Get(value.key)
	case positionQuery:
		return ctx.Proxy().URI().GetQuery(value.key)
	case positionBody:

		if contentType == "application/x-www-form-urlencoded" {
			if !value.optional {
				return ctx.Proxy().Body().GetForm(value.key)
			}
			form, _ := ctx.Proxy().Body().BodyForm()
			if _, ok := form[value.key]; ok {
				return value.key
			}
		} else if contentType == "application/json" {
			key := value.key
			if !strings.HasPrefix(key, "$.") {
				key = "$." + key
			}

			x, err := jp.ParseString(key)
			if err != nil {
				log.Errorf("parse json path(%s) error: %v", key, err)
				return ""
			}
			result := x.Get(body)

			if len(result) > 0 {
				if value.optional {
					return value.key
				}

				switch r := result[0].(type) {
				case string:
					return r
				case float32, float64:
					return fmt.Sprintf("%.f", r)
				case bool:
					return strconv.FormatBool(r)
				default:
					return oj.JSON(r)
				}
			}
		}
	case positionSystem:
		return ctx.GetLabel(value.key)
	}
	return ""
}
