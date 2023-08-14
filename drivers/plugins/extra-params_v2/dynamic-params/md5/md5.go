package md5

import (
	"errors"
	"strings"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/utils"

	dynamic_params "github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type MD5 struct {
	*dynamic_params.Param
}

func NewMD5(name string, value []string) *MD5 {
	return &MD5{
		Param: dynamic_params.NewParam(name, value),
	}
}

func (m *MD5) Generate(ctx http_service.IHttpContext, contentType string, args ...interface{}) (interface{}, error) {
	result, err := m.Param.Generate(ctx, contentType, args...)
	if err != nil {
		return nil, err
	}
	v, ok := result.(string)
	if !ok {
		return nil, errors.New("md5 value is not string")
	}
	md5Value := utils.Md5(v)
	if !strings.HasPrefix(m.Name(), "__") {
		md5Value = strings.ToUpper(md5Value)
	}
	log.DebugF("md5 value before: %s,after: %s", v, md5Value)
	return md5Value, nil
}
