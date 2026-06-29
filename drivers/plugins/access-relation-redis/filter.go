package access_relation_redis

import (
	"fmt"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
)

func (w *AccessRelationRedis) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) (err error) {
	if len(w.rules) == 0 {
		return next.DoChain(ctx)
	}

	var cache resources.ICache
	var cl []resources.ICache
	if w.redisID != "" {
		cl = scope_manager.Auto[resources.ICache](string(w.redisID), "redis").List()
	}
	if len(cl) == 0 {
		cl = scope_manager.Get[resources.ICache]("redis").List()
	}
	if len(cl) > 0 {
		cache = cl[0]
	} else {
		w.response.Response(ctx)
		return fmt.Errorf("cache not found")
	}

	for _, rule := range w.rules {
		if w.checkRule(ctx, cache, rule) {
			return next.DoChain(ctx)
		}
	}

	w.response.Response(ctx)
	return nil
}

func (w *AccessRelationRedis) checkRule(ctx http_context.IHttpContext, cache resources.ICache, rule *ruleHandler) bool {
	redisKey := rule.redisKeyGenerator.Key(ctx)
	resourceID := rule.labelGenerator.Key(ctx)
	if redisKey == "" || resourceID == "" {
		return false
	}

	var err error
	ok, err := cache.SIsMember(ctx.Context(), redisKey, resourceID).Result()
	if err != nil {
		log.Errorf("[access-relation-redis]redis sismember error: %v", err)
		return false
	}
	return ok

	//
	//dataStr, err = cache.Get(ctx.Context(), redisKey).Result()
	//if err != nil {
	//	if errors.Is(err, redis.Nil) {
	//		return false
	//	}
	//	log.Errorf("[access-relation-redis]redis get error: %v", err)
	//	return false
	//}
	//
	//if err != nil || dataStr == "" {
	//	return false
	//}
	//
	//var resources []string
	//if err := json.Unmarshal([]byte(dataStr), &resources); err != nil {
	//	return false
	//}
	//
	//for _, r := range resources {
	//	if r == resourceID {
	//		return true
	//	}
	//}
	//
	//return false
}

func (w *AccessRelationRedis) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_context.DoHttpFilter(w, ctx, next)
}
