package model_mapping

import (
	"testing"

	"github.com/eolinker/apinto/strategy"
)

func TestCheck(t *testing.T) {
	// 1. 合法配置
	validCfg := &Config{
		Name:     "strat-1",
		Priority: 10,
		Filters: strategy.FilterConfig{
			"consumer": []string{"c1"},
			"model":    []string{"claude-opus-5"},
		},
		Targets: []string{"Anthropic/claude-opus-5", "Anthropic-Pro/claude-opus-5"},
	}
	if _, err := Check(validCfg); err != nil {
		t.Fatalf("expected valid config, got err: %v", err)
	}

	// 2. Priority 越界
	invalidPriority := *validCfg
	invalidPriority.Priority = 0
	if _, err := Check(&invalidPriority); err == nil {
		t.Fatal("expected error for priority 0, got nil")
	}

	// 3. Targets 格式非法
	invalidTarget := *validCfg
	invalidTarget.Targets = []string{"invalid_target_without_slash"}
	if _, err := Check(&invalidTarget); err == nil {
		t.Fatal("expected error for invalid target format, got nil")
	}

	// 4. 缺少 consumer 或 model 过滤
	noConsumer := *validCfg
	noConsumer.Filters = strategy.FilterConfig{
		"model": []string{"claude-opus-5"},
	}
	if _, err := Check(&noConsumer); err == nil {
		t.Fatal("expected error when consumer filter is missing, got nil")
	}

	noModel := *validCfg
	noModel.Filters = strategy.FilterConfig{
		"consumer": []string{"c1"},
	}
	if _, err := Check(&noModel); err == nil {
		t.Fatal("expected error when model filter is missing, got nil")
	}
}

func TestManager_MatchBest(t *testing.T) {
	mgr := NewManager("consumer", "model")

	cfg1 := &Config{
		Name:     "strat-high-priority",
		Priority: 10,
		Filters: strategy.FilterConfig{
			"consumer": []string{"consumer-A"},
			"model":    []string{"claude-opus-5", "claude-alias-fast"},
		},
		Targets: []string{"Anthropic/claude-opus-5", "Anthropic-Pro/claude-opus-5"},
	}
	h1, err := NewHandler(cfg1)
	if err != nil {
		t.Fatalf("create handler 1 error: %v", err)
	}

	cfg2 := &Config{
		Name:     "strat-low-priority",
		Priority: 20,
		Filters: strategy.FilterConfig{
			"consumer": []string{"consumer-A"},
			"model":    []string{"claude-opus-5"},
		},
		Targets: []string{"Bedrock/claude-opus-5"},
	}
	h2, err := NewHandler(cfg2)
	if err != nil {
		t.Fatalf("create handler 2 error: %v", err)
	}

	mgr.Set("strat-1", h1, cfg1.Filters)
	mgr.Set("strat-2", h2, cfg2.Filters)

	// 1. 匹配 consumer-A + claude-opus-5，期望命中 Priority=10 的 h1
	ctxMatch := &mockContext{labels: map[string]string{
		"consumer": "consumer-A",
		"model":    "claude-opus-5",
	}}
	best, ok := mgr.MatchBest(ctxMatch)
	if !ok || best.Name() != "strat-high-priority" {
		t.Fatalf("expected strat-high-priority, got ok=%v, name=%s", ok, best.Name())
	}
	if best.PrimaryTarget().Raw != "Anthropic/claude-opus-5" {
		t.Fatalf("expected primary target Anthropic/claude-opus-5, got %s", best.PrimaryTarget().Raw)
	}
	if len(best.FailoverTargets()) != 1 || best.FailoverTargets()[0].Raw != "Anthropic-Pro/claude-opus-5" {
		t.Fatalf("expected 1 failover target Anthropic-Pro/claude-opus-5, got %+v", best.FailoverTargets())
	}

	// 2. 匹配笛卡尔积展开的第二个模型别名 claude-alias-fast
	ctxFast := &mockContext{labels: map[string]string{
		"consumer": "consumer-A",
		"model":    "claude-alias-fast",
	}}
	bestFast, ok := mgr.MatchBest(ctxFast)
	if !ok || bestFast.Name() != "strat-high-priority" {
		t.Fatalf("expected hit strat-high-priority for fast alias, got %v", ok)
	}

	// 3. 上下文缺少 consumer 或 model：无兜底，立即返回 false
	ctxNoConsumer := &mockContext{labels: map[string]string{"model": "claude-opus-5"}}
	if _, ok := mgr.MatchBest(ctxNoConsumer); ok {
		t.Fatal("expected no match when consumer missing")
	}

	ctxNoModel := &mockContext{labels: map[string]string{"consumer": "consumer-A"}}
	if _, ok := mgr.MatchBest(ctxNoModel); ok {
		t.Fatal("expected no match when model missing")
	}

	// 4. 未匹配的 consumer 或 model：无兜底返回 false
	ctxMismatch := &mockContext{labels: map[string]string{
		"consumer": "consumer-B",
		"model":    "claude-opus-5",
	}}
	if _, ok := mgr.MatchBest(ctxMismatch); ok {
		t.Fatal("expected no match for consumer-B")
	}

	// 5. 删除高优先级策略 strat-1，再次匹配应命中 strat-2
	mgr.Del("strat-1")
	bestAfterDel, ok := mgr.MatchBest(ctxMatch)
	if !ok || bestAfterDel.Name() != "strat-low-priority" {
		t.Fatalf("expected strat-low-priority after del strat-1, got ok=%v", ok)
	}
}
