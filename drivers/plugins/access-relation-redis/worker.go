package access_relation_redis

import (
	"github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"net/http"
)

var (
	_ eocontext.IFilter       = (*AccessRelationRedis)(nil)
	_ http_context.HttpFilter = (*AccessRelationRedis)(nil)
	_ eosc.IWorker            = (*AccessRelationRedis)(nil)
)

type AccessRelationRedis struct {
	drivers.WorkerBase
	redisID  eosc.RequireId
	rules    []*ruleHandler
	response response.IResponse
}

type ruleHandler struct {
	redisKeyGenerator context_label.IKeyGenerator
	labelGenerator    context_label.IKeyGenerator
}

func (w *AccessRelationRedis) Start() error {
	return nil
}

func (w *AccessRelationRedis) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	config, err := assert(conf)
	if err != nil {
		return err
	}
	err = Check(config, workers)
	if err != nil {
		return err
	}
	return w.reset(config)
}

func (w *AccessRelationRedis) reset(config *Config) error {
	w.redisID = config.Cache
	responseHandler := response.Parse(config.Response)
	if responseHandler == nil {
		responseHandler = defaultResponse
	}
	w.response = responseHandler

	rules := make([]*ruleHandler, 0, len(config.Rules))
	for _, r := range config.Rules {
		rules = append(rules, &ruleHandler{
			redisKeyGenerator: context_label.NewKeyGenerator(r.RedisKey),
			labelGenerator:    context_label.NewKeyGenerator(r.Label),
		})
	}
	w.rules = rules
	return nil
}

func (w *AccessRelationRedis) Stop() error {
	return nil
}

func (w *AccessRelationRedis) Destroy() {
}

func (w *AccessRelationRedis) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}

func assert(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	return cfg, nil
}

var (
	defaultResponse = response.Parse(&response.Response{
		StatusCode:  http.StatusForbidden,
		ContentType: "text/plain",
		Charset:     "utf-8",
		Headers:     nil,
		Body:        http.StatusText(http.StatusForbidden),
	})
)
