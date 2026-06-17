package variable_relation

import (
	"net/http"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/apinto/utils/context-label"
	"github.com/eolinker/apinto/utils/response"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

var (
	_ eocontext.IFilter       = (*VariableRelation)(nil)
	_ http_context.HttpFilter = (*VariableRelation)(nil)
	_ eosc.IWorker            = (*VariableRelation)(nil)
)

// RuleHandler 关系规则处理器
type RuleHandler struct {
	keyGenerator context_label.IKeyGenerator
	innerKeyTpl  []Segment
	valueLabel   string
}

// VariableRelation 变量关系插件Worker
type VariableRelation struct {
	drivers.WorkerBase
	rules    []*RuleHandler
	response response.IResponse
}

func (w *VariableRelation) Start() error {
	return nil
}

func (w *VariableRelation) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	config, err := assert(conf)
	if err != nil {
		return err
	}
	err = Check(config, workers)
	if err != nil {
		return err
	}
	return w.parseConfig(config)
}

func (w *VariableRelation) Stop() error {
	return nil
}

func (w *VariableRelation) Destroy() {
}

func (w *VariableRelation) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
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

func (w *VariableRelation) parseConfig(config *Config) error {
	responseHandler := response.Parse(config.Response)
	if responseHandler == nil {
		responseHandler = defaultResponse
	}
	w.response = responseHandler

	rules := make([]*RuleHandler, 0, len(config.Rules))
	for _, r := range config.Rules {
		keyGen := context_label.NewKeyGenerator(r.Key)
		tpl := ParseTemplate(r.InnerKey)
		valLabel := extractLabelName(r.ValueLabel)

		rules = append(rules, &RuleHandler{
			keyGenerator: keyGen,
			innerKeyTpl:  tpl,
			valueLabel:   valLabel,
		})
	}
	w.rules = rules
	return nil
}

func assert(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	return cfg, nil
}

// extractLabelName 从形如 "{api}" 的值模式中提取标签名称 "api"
func extractLabelName(template string) string {
	if len(template) > 2 && template[0] == '{' && template[len(template)-1] == '}' {
		return template[1 : len(template)-1]
	}
	return template
}
