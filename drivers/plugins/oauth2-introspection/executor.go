package oauth2_introspection

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/http"
	"sync"
	"time"
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	client         http.Client
	endpoint       string
	clientId       string
	clientSecret   string
	tokenName      string
	scopes         map[string]struct{}
	ttl            time.Duration
	claims         []string
	consumeBy      string
	hideCredential bool
	allowAnonymous bool
	once           sync.Once
	cache          scope_manager.IProxyOutput[resources.ICache]
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) reset(conf *Config) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	if !conf.IntrospectionSSLVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	e.hideCredential = conf.HideCredential
	e.client = client
	e.endpoint = conf.IntrospectionEndpoint
	e.clientId = conf.ClientID
	e.clientSecret = conf.ClientSecret
	e.tokenName = conf.TokenHeader
	e.scopes = make(map[string]struct{})
	for _, scope := range conf.Scopes {
		e.scopes[scope] = struct{}{}
	}
	e.ttl = time.Duration(conf.TTL) * time.Second
	e.claims = conf.CustomClaimsForward
	e.consumeBy = conf.ConsumerBy
	e.allowAnonymous = conf.AllowAnonymous

	return nil

}

func (e *executor) Stop() error {
	e.Destroy()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) Destroy() {
	e.client.CloseIdleConnections()
	return
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	token := retrieveAccessToken(ctx, positionHeader, e.tokenName)
	if token == "" {
		ctx.Response().SetBody([]byte("empty token"))
		ctx.Response().SetStatus(http.StatusUnauthorized, "empty token")
		return fmt.Errorf("empty token")
	}
	e.once.Do(func() {
		e.cache = scope_manager.Auto[resources.ICache]("", "redis")
	})
	ctx.SetLabel("token", token)
	var introspectionInfo *eosc.Base[IntrospectionResponseBody]
	var cache resources.ICache
	if len(e.cache.List()) > 0 {
		cache = e.cache.List()[0]
	}
	if cache != nil {
		d, err := cache.Get(ctx.Context(), fmt.Sprintf("%s:%s", redisKeyPrefix, token)).Result()
		if err == nil {
			var t eosc.Base[IntrospectionResponseBody]
			err = json.Unmarshal([]byte(d), &t)
			if err == nil {
				introspectionInfo = &t
			}
		}
	}
	if (introspectionInfo != nil && !checkActive(introspectionInfo.Config)) || introspectionInfo == nil {
		// 当缓存信息不存在或者缓存信息过期时，重新发起请求
		introspectionInfo, err = doIntrospectAccessToken(&e.client, e.endpoint, e.clientId, e.clientSecret, token)
		if err != nil {
			errInfo := fmt.Sprintf("do introspect access token error: %s", err.Error())
			ctx.Response().SetBody([]byte(errInfo))
			ctx.Response().SetStatus(http.StatusInternalServerError, "Internal Server Error")
			return fmt.Errorf(errInfo)
		}
	}
	err = verifyIntrospection(introspectionInfo.Config, e.scopes)
	if err != nil {
		// 校验失败
		errInfo := fmt.Sprintf("verify introspection error: %s", err.Error())
		ctx.Response().SetBody([]byte(errInfo))
		ctx.Response().SetStatus(http.StatusUnauthorized, "Unauthorized")
		return fmt.Errorf(errInfo)
	}
	err = setAppLabel(ctx, introspectionInfo.Config, e.consumeBy, e.allowAnonymous)
	if err != nil {
		errInfo := fmt.Sprintf("set app label error: %s", err.Error())
		ctx.Response().SetBody([]byte(errInfo))
		ctx.Response().SetStatus(http.StatusUnauthorized, "Unauthorized")
		return fmt.Errorf(errInfo)
	}

	if cache != nil {
		d, err := json.Marshal(introspectionInfo)
		if err == nil {
			_, err = cache.SetNX(ctx.Context(), fmt.Sprintf("%s:%s", redisKeyPrefix, token), d, e.ttl).Result()
			if err != nil {
				errInfo := fmt.Sprintf("set cache error: %s", err.Error())
				ctx.Response().SetBody([]byte(errInfo))
				ctx.Response().SetStatus(http.StatusInternalServerError, "Internal Server Error")
				return fmt.Errorf(errInfo)
			}
		}
	}
	if e.hideCredential {
		ctx.Proxy().Header().DelHeader(e.tokenName)
	}

	ctx.Proxy().Header().SetHeader("X-Credential-Scope", introspectionInfo.Config.Scope)
	ctx.Proxy().Header().SetHeader("X-Credential-Client-ID", introspectionInfo.Config.ClientId)
	ctx.Proxy().Header().SetHeader("X-Credential-Token-Type", "Bearer")
	ctx.Proxy().Header().SetHeader("X-Credential-Exp", fmt.Sprintf("%d", introspectionInfo.Config.Exp))
	ctx.Proxy().Header().SetHeader("X-Credential-Iat", fmt.Sprintf("%d", introspectionInfo.Config.Iat))
	ctx.Proxy().Header().SetHeader("X-Credential-Nbf", fmt.Sprintf("%d", introspectionInfo.Config.Nbf))
	ctx.Proxy().Header().SetHeader("X-Credential-Sub", introspectionInfo.Config.Sub)
	ctx.Proxy().Header().SetHeader("X-Credential-Aud", introspectionInfo.Config.Aud)
	ctx.Proxy().Header().SetHeader("X-Credential-Iss", introspectionInfo.Config.Iss)
	ctx.Proxy().Header().SetHeader("X-Credential-Jti", introspectionInfo.Config.Jti)
	for _, v := range e.claims {
		a, ok := introspectionInfo.Append[v]
		if !ok {
			continue
		}
		vv, ok := a.(string)
		if !ok {
			continue
		}
		ctx.Proxy().Header().SetHeader(fmt.Sprintf("X-Credential-%s", v), vv)
	}

	if next != nil {
		return next.DoChain(ctx)
	}

	return nil
}
