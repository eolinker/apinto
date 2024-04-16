package access_relational

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	"github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/metrics"
	"net/http"
)

var (
	_ eocontext.IFilter       = (*AccessRelational)(nil)
	_ http_context.HttpFilter = (*AccessRelational)(nil)
	_ eosc.IWorker            = (*AccessRelational)(nil)
)

type AccessRelational struct {
	drivers.WorkerBase
	data     eosc.ICustomerVar
	rules    []*ruleHandler
	response response.IResponse
}

type ruleHandler struct {
	key   metrics.Metrics
	field metrics.Metrics
}

func (a *AccessRelational) Start() error {
	return nil
}

func (a *AccessRelational) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	config, err := assert(conf)
	if err != nil {
		return err
	}
	err = Check(config, workers)
	if err != nil {
		return err
	}
	iResponse, handlers := parseConfig(config)
	a.response = iResponse
	a.rules = handlers
	return nil
}

func (a *AccessRelational) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (a *AccessRelational) Destroy() {
	//TODO implement me
	panic("implement me")
}
func (a *AccessRelational) CheckSkill(skill string) bool {
	//TODO implement me
	panic("implement me")
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
		Body:        `403 Forbidden`,
	})
)

func parseConfig(config *Config) (response.IResponse, []*ruleHandler) {

	responseHandler := response.Parse(config.Response)
	if responseHandler == nil {
		responseHandler = defaultResponse
	}
	rules := make([]*ruleHandler, 0)
	for _, rule := range config.Rules {
		key := metrics.Parse(rule.KeyRule)
		field := metrics.Parse(rule.AccessRule)
		if key == nil || field == nil {
			continue
		}
		rules = append(rules, &ruleHandler{
			key:   key,
			field: field,
		})
	}
	return responseHandler, rules
}
