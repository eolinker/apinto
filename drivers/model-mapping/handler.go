package model_mapping

import (
	"fmt"

	"github.com/eolinker/apinto/strategy"
)

// IHandler 模型别名映射策略处理器接口
type IHandler interface {
	Name() string
	Priority() int
	PrimaryTarget() *Target
	Targets() []*Target
	FailoverTargets() []*Target
	Filters() strategy.FilterConfig
}

type handler struct {
	name            string
	priority        int
	primaryTarget   *Target
	targets         []*Target
	failoverTargets []*Target
	filters         strategy.FilterConfig
}

// NewHandler 根据配置创建策略 Handler
func NewHandler(cfg *Config) (IHandler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	targets := make([]*Target, 0, len(cfg.Targets))
	for _, tStr := range cfg.Targets {
		t, err := ParseTarget(tStr)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("targets cannot be empty")
	}

	primary := targets[0]
	var failovers []*Target
	if len(targets) > 1 {
		failovers = targets[1:]
	}

	return &handler{
		name:            cfg.Name,
		priority:        cfg.Priority,
		primaryTarget:   primary,
		targets:         targets,
		failoverTargets: failovers,
		filters:         cfg.Filters,
	}, nil
}

func (h *handler) Name() string {
	return h.name
}

func (h *handler) Priority() int {
	return h.priority
}

func (h *handler) PrimaryTarget() *Target {
	return h.primaryTarget
}

func (h *handler) Targets() []*Target {
	return h.targets
}

func (h *handler) FailoverTargets() []*Target {
	return h.failoverTargets
}

func (h *handler) Filters() strategy.FilterConfig {
	return h.filters
}
