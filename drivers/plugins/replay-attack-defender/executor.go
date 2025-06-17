package replay_attack_defender

import (
	"errors"
	"fmt"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/eolinker/apinto/utils"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var _ http_service.HttpFilter = (*executor)(nil)
var _ eocontext.IFilter = (*executor)(nil)
var _ eosc.IWorker = (*executor)(nil)

type executor struct {
	drivers.WorkerBase
	signHeader      string
	timestampHeader string
	nonceHeader     string
	token           string
	ttl             time.Duration
	cache           scope_manager.IProxyOutput[resources.ICache]
	once            sync.Once
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	return http_service.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_service.IHttpContext, next eocontext.IChain) (err error) {
	e.once.Do(func() {
		e.cache = scope_manager.Auto[resources.ICache]("", "redis")
	})
	timestamp, err := checkTimestamp(ctx.Request().Header().GetHeader(e.timestampHeader))
	if err != nil {
		ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
		ctx.Response().SetBody([]byte(err.Error()))
		return err
	}
	
	nonce := ctx.Request().Header().GetHeader(e.nonceHeader)
	if nonce == "" {
		ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
		ctx.Response().SetBody([]byte("missing nonce"))
		return fmt.Errorf("missing nonce")
	}
	
	sign := ctx.Request().Header().GetHeader(e.signHeader)
	if sign == "" {
		ctx.Response().SetStatus(http.StatusBadRequest, "Bad Request")
		ctx.Response().SetBody([]byte("missing sign"))
		return fmt.Errorf("missing sign")
	}
	signBefore := fmt.Sprintf("%d%s%s", timestamp, nonce, e.token)
	signAfter := utils.Md5(signBefore)
	if sign != signAfter {
		ctx.Response().SetStatus(http.StatusForbidden, "Forbidden")
		ctx.Response().SetBody([]byte("invalid sign"))
		return fmt.Errorf("invalid sign")
	}
	key := "apinto:replay-attack-defender:" + sign
	for _, c := range e.cache.List() {
		if c == nil {
			continue
		}
		_, err = c.Get(ctx.Context(), key).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				c.SetNX(ctx.Context(), key, []byte(signBefore), e.ttl)
				break
			}
			if err.Error() != "redis: nil" {
				continue
			}
		}
		err = fmt.Errorf("replay attack detected, nonce: %s, sign: %s, timestamp: %d", nonce, sign, timestamp)
		ctx.Response().SetStatus(http.StatusForbidden, "Forbidden")
		ctx.Response().SetBody([]byte(err.Error()))
		return err
	}
	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func (e *executor) Destroy() {
	return
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	return nil
}

func (e *executor) reset(conf *Config) error {
	
	if conf.NonceHeader == "" {
		conf.NonceHeader = "X-Ca-Nonce"
	}
	if conf.SignHeader == "" {
		conf.SignHeader = "X-Ca-Signature"
	}
	if conf.TimestampHeader == "" {
		conf.TimestampHeader = "X-Ca-Timestamp"
	}
	if conf.TTL < 1 {
		conf.TTL = 600
	}
	e.signHeader = conf.SignHeader
	e.timestampHeader = conf.TimestampHeader
	e.nonceHeader = conf.NonceHeader
	e.token = conf.ReplayAttackToken
	e.ttl = time.Duration(conf.TTL) * time.Second
	return nil
}

func (e *executor) Stop() error {
	e.Destroy()
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_service.FilterSkillName == skill
}

func checkTimestamp(timestamp string, ) (int64, error) {
	if timestamp == "" {
		return 0, fmt.Errorf("missing timestamp")
	}
	
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp %s, error: %v", timestamp, err)
	}
	now := time.Now()
	t := time.Unix(ts, 0)
	if now.Sub(t) > 10*time.Minute {
		return 0, fmt.Errorf("timestamp %d is too old, current time is %d", ts, now.Unix())
	} else if now.Sub(t) < -10*time.Minute {
		return 0, fmt.Errorf("timestamp %d is too new, current time is %d", ts, now.Unix())
	}
	return ts, nil
}
