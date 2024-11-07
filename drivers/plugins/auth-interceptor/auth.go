package auth_interceptor

import (
	"encoding/json"
	"fmt"
	"mime"
	"time"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type Auth struct {
	cfg       *Config
	redisConn string
	drivers.WorkerBase
}

func (a *Auth) Destroy() {
	redisPool.Release(a.redisConn)
	return
}

// DoFilter 拦截请求过滤，内部转换为http类型再处理
func (a *Auth) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(a, ctx, next)
}

// Start 插件添加时执行
func (a *Auth) Start() error {
	return nil
}

// Reset 当类型为插件时，Reset方法不会执行
func (a *Auth) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

// Stop 插件删除时间执行
func (a *Auth) Stop() error {
	return nil
}

func (a *Auth) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

// DoHttpFilter 核心http处理方法
func (a *Auth) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) error {

	err := a.addAuthToProxy(ctx)
	if err != nil {
		ctx.Response().SetStatus(401, "Unauthorized")
		ctx.Response().SetBody([]byte(err.Error()))
		return err
	}
	err = next.DoChain(ctx)
	if err != nil {
		return err
	}
	if ctx.Response().StatusCode() == 401 {
		client := redisPool.Get(a.redisConn)
		client.Del(ctx.Context(), a.cfg.SysKey).Result()
		err = a.addAuthToProxy(ctx)
		if err != nil {
			ctx.Response().SetStatus(401, "Unauthorized")
			ctx.Response().SetBody([]byte(err.Error()))
			return err
		}
		return next.DoChain(ctx)
	}
	return nil
}

func (a *Auth) addAuthToProxy(ctx http_service.IHttpContext) error {
	var token string
	var err error
	client := redisPool.Get(a.redisConn)

	for i := 0; i < a.cfg.RetryCount+1; i++ {
		token, err = client.Get(ctx.Context(), a.cfg.SysKey).Result()
		if err == nil {
			break
		}
		time.Sleep(time.Duration(a.cfg.RetryPeriod) * time.Second)
	}
	if token == "" || err != nil {
		client.Del(ctx.Context(), a.cfg.SysKey)
		return fmt.Errorf("[plugin auth-interceptor invoke err] get token failed")
	}

	switch a.cfg.AuthPosition {
	case positionHeader:
		ctx.Proxy().Header().SetHeader(a.cfg.AuthKey, token)
	case positionBody:
		// 如果是json类型的body，需要解析后再添加auth
		contentType, _, _ := mime.ParseMediaType(ctx.Proxy().Body().ContentType())
		switch contentType {
		case "application/x-www-form-urlencoded", "multipart/form-data":
			ctx.Proxy().Body().SetToForm(a.cfg.AuthKey, token)
		case "application/json":
			var body = make(map[string]interface{})
			bytes, _ := ctx.Proxy().Body().RawBody()
			err = json.Unmarshal(bytes, &body)
			if err != nil {
				return fmt.Errorf("[plugin auth-interceptor invoke err] unmarshal body failed")
			}
			body[a.cfg.AuthKey] = token
			rebody, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("[plugin auth-interceptor invoke err] marshal body failed")
			} else {
				ctx.Proxy().Body().SetRaw(ctx.Request().ContentType(), rebody)
			}
		}
	case positionQuery:
		ctx.Proxy().URI().SetQuery(a.cfg.AuthKey, token)
	}
	return nil
}

//
//// 添加auth头
//func AddAuth(a *Auth, ctx http_service.IHttpContext) {
//	var token, err = GetAuth(a.cfg)
//	// 如果没有认证字符串或错误，删除认证字符串，等待远程重新获取认证字符串
//	if token == "" && err != nil {
//		removeAuth(a.cfg)
//	}
//	// 轮询3次获取认证字符串，每次等待3秒
//	var retryTimes = 3
//	for i := 0; i < retryTimes; i++ {
//		time.Sleep(time.Second * 3)
//		token, err = GetAuth(a.cfg)
//		if token != "" && err == nil {
//			break
//		}
//	}
//	// 如果最后获取到了token,执行token附加流程
//	if token != "" {
//		switch a.cfg.AuthPosition {
//		case "header":
//			ctx.Proxy().Header().SetHeader(a.cfg.AuthKey, token)
//		case "body":
//			if strings.Contains(strings.ToLower(ctx.Request().ContentType()), "application/json") {
//				var body = make(map[string]interface{})
//
//				bytes, err := ctx.Request().Body().RawBody()
//				if err != nil {
//					json.Unmarshal(bytes, &body)
//				}
//				body[a.Config.AuthKey] = token
//				rebody, err := json.Marshal(body)
//				if err != nil {
//					ctx.Proxy().Body().SetRaw(ctx.Request().ContentType(), rebody)
//				}
//			} else {
//				ctx.Proxy().Body().AddForm(a.cfg.AuthKey, token)
//			}
//		case "query":
//			ctx.Proxy().URI().AddQuery(a.cfg.AuthKey, token)
//		}
//	}
//}
