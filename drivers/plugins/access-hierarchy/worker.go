package access_hierarchy

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
	// 确保 AccessHierarchy 实现了 http_context.HttpFilter (HTTP 过滤器) 和 eosc.IWorker (基础 Worker 实例) 接口
	_ eocontext.IFilter       = (*AccessHierarchy)(nil)
	_ http_context.HttpFilter = (*AccessHierarchy)(nil)
	_ eosc.IWorker            = (*AccessHierarchy)(nil)
)

// AccessHierarchy 实现多级渠道鉴权核心 Worker 实体
type AccessHierarchy struct {
	drivers.WorkerBase                    // 继承 Apinto 基础 Worker 结构
	rules              []ruleHandler      // 解析配置生成的各校验规则处理器
	response           response.IResponse // 校验失败时的 HTTP 响应模板
}

// Start 启动 Worker（本插件无需额外后台驻留协程，直接返回 nil）
func (w *AccessHierarchy) Start() error {
	return nil
}

// Reset 在插件配置被修改（更新）时被系统自动调用，重构规则处理器与响应模板
func (w *AccessHierarchy) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
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

// Stop 停止 Worker
func (w *AccessHierarchy) Stop() error {
	return nil
}

// Destroy 销毁 Worker，回收底层资源
func (w *AccessHierarchy) Destroy() {
}

// CheckSkill 判断当前 Worker 是否支持对应的 Skill（本插件只应用在 Http 过滤链中）
func (w *AccessHierarchy) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}

func assert(v interface{}) (*Config, error) {
	cfg, ok := v.(*Config)
	if !ok {
		return nil, eosc.ErrorConfigType
	}
	return cfg, nil
}

// 默认的 403 Forbidden 拦截响应模板
var (
	defaultResponse = response.Parse(&response.Response{
		StatusCode:  http.StatusForbidden,
		ContentType: "text/plain",
		Charset:     "utf-8",
		Headers:     nil,
		Body:        http.StatusText(http.StatusForbidden),
	})
)

// newHandler 为指定的 A、B 指标生成校验处理器
func (w *AccessHierarchy) newHandler(a, b string) ruleHandler {
	am := metrics.Parse(a) // 提取 A（通常是 AppID 指标）
	bm := metrics.Parse(b) // 提取 B（通常是 ResourceID 指标）
	if am == nil || bm == nil {
		return nil
	}
	return &handler{
		a: am,
		b: bm,
	}
}

// parseConfig 转换配置项，初始化规则和响应拦截器
func (w *AccessHierarchy) parseConfig(config *Config) (response.IResponse, []ruleHandler) {
	responseHandler := response.Parse(config.Response)
	if responseHandler == nil {
		responseHandler = defaultResponse // 若未定义自定义 Response，则降级到 403
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
