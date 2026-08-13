package quota_limiting_strategy

import (
	"testing"

	"github.com/eolinker/apinto/drivers/strategy"
	"github.com/eolinker/eosc"
)

func TestExecutorLifecycle(t *testing.T) {
	extractorManager = NewExtractor()
	controller = strategy.NewController("strategy", "quota-limiting", configType)

	cfg := &Config{
		Na: "strat_exec_1",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 100},
		},
		Filters: FiltersConfig{
			Tenant:   "tenant_exec",
			Resource: Filter{All: true},
			Target:   Filter{Type: "all", All: true},
		},
	}

	worker, err := Create("worker_1", "strat_exec_1", cfg, nil)
	if err != nil {
		t.Fatalf("expected create executor success, got %v", err)
	}

	exec, ok := worker.(*executor)
	if !ok {
		t.Fatalf("expected worker to be *executor")
	}

	if err := exec.Start(); err != nil {
		t.Errorf("expected Start nil, got %v", err)
	}

	if exec.CheckSkill("any_skill") != false {
		t.Errorf("expected CheckSkill false")
	}

	// 测试 Reset 非 Config 类型参数
	if err := exec.Reset("invalid_config_type", nil); err != eosc.ErrorConfigType {
		t.Errorf("expected ErrorConfigType, got %v", err)
	}

	// 确认策略已添加到 extractorManager
	ctx := &dummyContext{
		labels: map[string]string{"tenant": "tenant_exec"},
	}
	if _, has := GetRequestStrategies(ctx); !has {
		t.Errorf("expected strategy added after Create")
	}

	// 测试 Stop
	if err := exec.Stop(); err != nil {
		t.Errorf("expected Stop nil, got %v", err)
	}

	if _, has := GetRequestStrategies(ctx); has {
		t.Errorf("expected strategy removed after Stop")
	}
}
