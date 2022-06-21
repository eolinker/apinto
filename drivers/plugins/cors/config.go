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
	AllowOrigins     string `json:"allow_origins" label:"允许跨域访问的Origin" default:"*"`
	AllowMethods     string `json:"allow_methods" label:"允许通过的请求方式" default:"*" description:"多种请求方式用英文逗号隔开"`
	AllowCredentials bool   `json:"allow_credentials" label:"请求中是否携带cookie"`
	AllowHeaders     string `json:"allow_headers" label:"允许跨域访问时请求方携带的非CORS规范以外的Header" default:"*" description:"多种请求方式用英文逗号隔开"`
	ExposeHeaders    string `json:"expose_headers" label:"允许跨域访问时响应方携带的非CORS规范以外的Header" default:"*" description:"多种请求方式用英文逗号隔开"`
	MaxAge           int32  `json:"max_age" description:"浏览器缓存CORS结果的最大时间" description:"单位：s，最小值：1" default:"5" minimum:"1"`
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
