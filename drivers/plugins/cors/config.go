package cors

import (
	"errors"
	http_service "github.com/eolinker/eosc/http-service"
	"strconv"
)

var (
	ErrorCredentialTypeError = errors.New("the other options cannot be `*` when `allowCredentials` is true")
)

type Config struct {
	AllowOrigins     string `json:"allow_origins"`
	AllowMethods     string `json:"allow_methods"`
	AllowCredentials bool   `json:"allow_credentials"`
	AllowHeaders     string `json:"allow_headers"`
	ExposeHeaders    string `json:"expose_headers"`
	MaxAge           int32  `json:"max_age"`
	ResponseType     string `json:"response_type"`
}

func (c *Config) doCheck() error {
	if c.AllowCredentials {
		// allowCredentials为true时其他项不能为*
		if c.AllowOrigins == "*" || c.AllowMethods == "*" || c.AllowHeaders == "*" || c.ExposeHeaders == "*" {
			return ErrorCredentialTypeError
		}
	}
	// 不填则填充默认值
	if c.AllowOrigins == "" {
		c.AllowOrigins = "*"
	}
	if c.AllowMethods == "" {
		c.AllowMethods = "*"
	}
	if c.AllowHeaders == "" {
		c.AllowHeaders = "*"
	}
	if c.ExposeHeaders == "" {
		c.ExposeHeaders = "*"
	}
	if c.MaxAge == 0 {
		c.MaxAge = 5
	}
	if c.ResponseType != "json" {
		c.ResponseType = "text"
	}
	return nil
}

type optionHandler func(ctx http_service.IHttpContext) error

func (c *Config) genOptionHandler() optionHandler {
	return func(ctx http_service.IHttpContext) error {
		info := "[CORS] Cross Domain!"
		resp := ctx.Response()
		ctx.Response().SetHeader("Access-Control-Allow-Origin", c.AllowOrigins)
		ctx.Response().SetHeader("Access-Control-Allow-Methods", c.AllowMethods)
		ctx.Response().SetHeader("Access-Control-Max-Age", string(c.MaxAge))
		ctx.Response().SetHeader("Access-Control-Expose-Headers", c.ExposeHeaders)
		ctx.Response().SetHeader("Access-Control-Allow-Headers", c.AllowHeaders)
		ctx.Response().SetHeader("Access-Control-Allow-Credentials", strconv.FormatBool(c.AllowCredentials))
		resp.SetStatus(200, "200")
		resp.SetBody([]byte(info))
		return nil
	}
}
