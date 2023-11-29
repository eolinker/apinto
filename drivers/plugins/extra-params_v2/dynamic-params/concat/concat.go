package concat

import (
	"errors"
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	dynamic_params "github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params"
)

type Config struct {
	*dynamic_params.Param
}

func (c *Config) Generate(ctx http_service.IHttpContext, contentType string, args ...interface{}) (interface{}, error) {
	result, err := c.Param.Generate(ctx, contentType, args...)
	if err != nil {
		return nil, err
	}
	v, ok := result.(string)
	if !ok {
		return nil, errors.New("concat value is not string")
	}
	if !strings.HasPrefix(c.Name(), "__") {
		v = strings.ToUpper(v)
	}
	return v, nil
}

func NewConcat(name string, value []string) *Config {
	return &Config{
		Param: dynamic_params.NewParam(name, value),
	}
}
