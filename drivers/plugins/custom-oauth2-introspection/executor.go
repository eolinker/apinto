package custom_oauth2_introspection

import (
	"errors"
	"fmt"
	"github.com/coocood/freecache"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/redis/go-redis/v9"
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	endpoint             string
	introspectionRequest *IntrospectionRequest
	tokenPosition        string
	tokenName            string
	tokenExpr            jp.Expr
	cache                scope_manager.IProxyOutput[resources.ICache]
	once                 sync.Once
	ttl                  int
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {

	return e.reset(conf.(*Config))
}

func (e *executor) reset(conf *Config) error {
	if e.tokenPosition == positionBody {
		tokenName := conf.TokenName
		if !strings.HasPrefix(tokenName, "$.") {
			tokenName = fmt.Sprintf("$.%s", tokenName)
		}
		expr, err := jp.ParseString(tokenName)
		if err != nil {
			return err
		}
		e.tokenExpr = expr
	}
	e.endpoint = conf.Endpoint
	e.introspectionRequest = conf.IntrospectionRequest
	e.tokenPosition = conf.TokenPosition
	e.tokenName = conf.TokenName
	e.ttl = conf.TTL
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
	e.cache = nil
	return
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	e.once.Do(func() {
		e.cache = scope_manager.Auto[resources.ICache]("", "redis")
	})
	api := ctx.GetLabel("api")
	app := ctx.GetLabel("application")
	key := fmt.Sprintf("%s:%s:%s", redisKeyPrefix, api, app)

	cache := e.pickCache()
	token, fromCache, err := e.loadOrFetchToken(ctx, cache, key)
	if err != nil {
		ctx.Response().SetStatus(401, "Unauthorized")
		ctx.Response().SetBody([]byte(err.Error()))
		return err
	}

	err = e.injectToProxy(ctx, token)
	if err != nil {
		return err
	}

	err = next.DoChain(ctx)
	if err != nil {
		return err
	}
	// unauthorized → purge and retry once
	if ctx.Response().StatusCode() == http.StatusUnauthorized || ctx.Response().StatusCode() == http.StatusForbidden {
		if fromCache {
			cache.Del(ctx.Context(), key)
		}
		token, err = e.fetchAndCacheToken(ctx, cache, key)
		if err != nil {
			ctx.Response().SetStatus(401, "Unauthorized")
			ctx.Response().SetBody([]byte(err.Error()))
			return err
		}
		err = e.injectToProxy(ctx, token)
		if err != nil {
			return err
		}
		return next.DoChain(ctx)
	}
	return nil
}

func (e *executor) fetchAndCacheToken(ctx http_service.IHttpContext, cache resources.ICache, key string) (string, error) {
	// lock on key
	lockKey := fmt.Sprintf("lock:%s", key)
	ok, err := cache.AcquireLock(ctx.Context(), lockKey, ctx.RequestId(), 10).Result()
	if err != nil {
		// fallback: no lock, proceed
		return "", err
	}
	if ok {
		defer cache.ReleaseLock(ctx.Context(), lockKey, ctx.RequestId())
	}
	token, err := cache.Get(ctx.Context(), key).Result()
	if err == nil {
		return token, nil
	}
	if err != nil {
		if !errors.Is(err, redis.Nil) && !errors.Is(err, freecache.ErrNotFound) {
			return "", err
		}
	}

	c := &IntrospectClient{
		Endpoint:        e.endpoint,
		Method:          e.introspectionRequest.Method,
		Header:          e.introspectionRequest.Header,
		Body:            e.introspectionRequest.Body,
		Query:           e.introspectionRequest.Query,
		ContentType:     e.introspectionRequest.ContentType,
		ExtractResponse: e.introspectionRequest.ExtractResponse,
	}
	var r *IntrospectResponse
	var lastErr error
	retry := e.introspectionRequest.Retry
	if retry <= 0 {
		retry = 1
	}
	for i := 0; i < retry; i++ {
		r, lastErr = c.Do(ctx)
		if lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	token = r.Token
	ttlSec := e.ttl
	if r.ExpiredIn > 0 && (ttlSec <= 0 || r.ExpiredIn < ttlSec) {
		ttlSec = r.ExpiredIn
	}
	if ttlSec <= 0 {
		ttlSec = 1
	} else {
		ttlSec = ttlSec - 5
		if ttlSec < 0 {
			ttlSec = 1
		}
	}

	// cache token mapping and also reverse mapping for additional params
	cache.Set(ctx.Context(), key, []byte(token), time.Duration(ttlSec)*time.Second)
	return token, nil
}

func (e *executor) loadOrFetchToken(ctx http_service.IHttpContext, cache resources.ICache, key string) (string, bool, error) {
	token, err := cache.Get(ctx.Context(), key).Result()
	if err == nil && token != "" {
		return token, true, nil
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		// fallback to local cache
		lc := resources.LocalCache()
		token, _ = lc.Get(ctx.Context(), key).Result()
		if token != "" {
			return token, true, nil
		}
	}
	token, err = e.fetchAndCacheToken(ctx, cache, key)
	if err != nil {
		// try local cache as last resort
		lc := resources.LocalCache()
		token2, err2 := e.fetchAndCacheToken(ctx, lc, key)
		if err2 != nil {
			return "", false, err2
		}
		return token2, false, nil
	}
	return token, false, nil
}

func (e *executor) injectToProxy(ctx http_service.IHttpContext, token string) error {
	switch e.tokenPosition {
	case positionHeader:
		ctx.Proxy().Header().SetHeader(e.tokenName, token)
	case positionQuery:
		ctx.Proxy().URI().SetQuery(e.tokenName, token)
	case positionBody:
		contentType := ctx.Proxy().Body().ContentType()
		switch {
		case strings.Contains(contentType, "application/x-www-form-urlencoded"), strings.Contains(contentType, "multipart/form-data"):
			return ctx.Proxy().Body().SetToForm(e.tokenName, token)
		case strings.Contains(contentType, "application/json"):
			body, _ := ctx.Proxy().Body().RawBody()
			o, err := oj.Parse(body)
			if err != nil {
				return err
			}
			return e.tokenExpr.Set(o, token)
		default:
			return fmt.Errorf("unsupported content type: %s", contentType)
		}
	default:
		return fmt.Errorf("unsupported token position: %s", e.tokenPosition)
	}
	return nil
}

func (e *executor) pickCache() resources.ICache {
	list := e.cache.List()
	if len(list) > 0 {
		return list[0]
	}
	return resources.LocalCache()
}
