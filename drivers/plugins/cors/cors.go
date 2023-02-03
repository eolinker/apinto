package cors

import (
	"encoding/json"
	"errors"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/http"
	"strconv"
	"strings"
)

var _ http_service.HttpFilter = (*CorsFilter)(nil)
var _ eocontext.IFilter = (*CorsFilter)(nil)

type CorsFilter struct {
	drivers.WorkerBase
	responseType     string
	allowCredentials bool
	option           optionHandler
	originChecker    *Checker
	methodChecker    *Checker
	headerChecker    *Checker
	exposeChecker    *Checker
}

func (c *CorsFilter) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(c, ctx, next)
}

func (c *CorsFilter) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	if ctx.Request().Method() == http.MethodOptions {
		return c.doOption(ctx)
	}
	err = c.doFilter(ctx)
	if err != nil {
		resp := ctx.Response()
		info := c.responseEncode(err.Error(), 403)
		resp.SetStatus(403, "403")
		resp.SetBody([]byte(info))
		return err
	}
	if next != nil {
		err = next.DoChain(ctx)
	}
	c.doNext(ctx)
	return err
}

func (c *CorsFilter) Destroy() {
	c.option = nil
	c.originChecker = nil
	c.methodChecker = nil
	c.headerChecker = nil
	c.exposeChecker = nil
	c.responseType = ""
}

func (c *CorsFilter) doOption(ctx http_service.IHttpContext) error {
	return c.option(ctx)
}

func (c *CorsFilter) doNext(ctx http_service.IHttpContext) {
	// 验证响应头部是否在expose-headers中
	for key := range ctx.Response().Headers() {
		if !c.exposeChecker.Check(key, true) {
			ctx.Response().DelHeader(key)
		}
	}
	c.WriteHeader(ctx)
}
func (c *CorsFilter) doFilter(ctx http_service.IHttpContext) error {
	check := ctx.Request().Header().GetHeader("Origin")
	// 验证源是否一致
	if !c.originChecker.Check(check, false) {
		// 头部反馈
		c.WriteHeader(ctx)
		// 结束
		resp := ctx.Response()
		info := "[CORS] The origin is not allowed"
		resp.SetStatus(400, "400")
		resp.SetBody([]byte(c.responseEncode(info, 400)))
		return errors.New(info)

	}
	check = ctx.Request().Method()
	// 验证请求方式是否允许
	if !c.methodChecker.Check(check, false) {
		// 头部反馈
		c.WriteHeader(ctx)
		// 结束
		resp := ctx.Response()
		info := "[CORS] Request method '" + ctx.Request().Method() + "' is not allowed"
		resp.SetStatus(400, "400")
		resp.SetBody([]byte(c.responseEncode(info, 400)))
		return errors.New(info)
	}
	// 验证自定义头部是否在allow-headers中
	for key := range ctx.Request().Header().Headers() {
		if !c.headerChecker.Check(key, true) {
			ctx.Proxy().Header().DelHeader(key)
		}
	}
	if !c.allowCredentials {
		cookie := ctx.Request().Header().GetHeader("Cookie")
		if cookie != "" {
			ctx.Proxy().Header().DelHeader("Cookie")
		}
	}
	return nil
}

// 全部匹配
func (c *CorsFilter) checkAllMatch(name string) bool {
	return strings.EqualFold(name, "*") || strings.EqualFold(name, "**")
}

func (c *CorsFilter) Start() error {
	return nil
}

func (c *CorsFilter) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := check(conf)
	if err != nil {
		return err
	}
	c.option = cfg.genOptionHandler()
	c.originChecker = NewChecker(cfg.AllowOrigins, "Access-Control-Allow-Origin")
	c.methodChecker = NewChecker(cfg.AllowMethods, "Access-Control-Allow-Methods")
	c.headerChecker = NewChecker(cfg.AllowHeaders, "Access-Control-Allow-Headers")
	c.exposeChecker = NewChecker(cfg.ExposeHeaders, "Access-Control-Expose-Headers")
	c.allowCredentials = cfg.AllowCredentials
	return nil
}

func (c *CorsFilter) Stop() error {
	return nil
}

func (c *CorsFilter) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

// WriteHeader CORS响应告诉本服务的规则
func (c *CorsFilter) WriteHeader(ctx http_service.IHttpContext) {
	resp := ctx.Response()
	c.writeHeader(resp, c.originChecker)
	c.writeHeader(resp, c.headerChecker)
	c.writeHeader(resp, c.methodChecker)
	c.writeHeader(resp, c.exposeChecker)
	resp.SetHeader("Access-Control-Allow-Credentials", strconv.FormatBool(c.allowCredentials))
}
func (c *CorsFilter) writeHeader(resp http_service.IResponse, h IHeader) {
	resp.SetHeader(h.GetKey(), h.GetOrigin())
}

func (c *CorsFilter) responseEncode(origin string, statusCode int) string {
	if c.responseType == "json" {
		tmp := map[string]interface{}{
			"message":     origin,
			"status_code": statusCode,
		}
		newInfo, _ := json.Marshal(tmp)
		return string(newInfo)
	}
	return origin
}
