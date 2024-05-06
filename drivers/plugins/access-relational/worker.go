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
	rules    []ruleHandler
	response response.IResponse
}

func (w *AccessRelational) Start() error {
	return nil
}

func (w *AccessRelational) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	config, err := assert(conf)
	if err != nil {
		return err
	}
	err = Check(config, workers)
	if err != nil {
		return err
	}
	iResponse, handlers := w.parseConfig(config)
	w.response = iResponse
	w.rules = handlers
	return nil
}

func (w *AccessRelational) Stop() error {
	return nil
}

func (w *AccessRelational) Destroy() {

}
func (w *AccessRelational) CheckSkill(skill string) bool {
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

func (w *AccessRelational) newHandler(a, b string) ruleHandler {
	am := metrics.Parse(a)
	bm := metrics.Parse(b)
	if am == nil || bm == nil {
		return nil
	}
	return &handler{
		a: am,
		b: bm,
	}
}
func (w *AccessRelational) parseConfig(config *Config) (response.IResponse, []ruleHandler) {

	responseHandler := response.Parse(config.Response)
	if responseHandler == nil {
		responseHandler = defaultResponse
	}
	rules := make([]ruleHandler, 0)
	for _, rule := range config.Rules {
		rh := w.newHandler(rule.A, rule.B)
		if rh != nil {
			rules = append(rules, rh)
		}
	}
	return responseHandler, rules
}
