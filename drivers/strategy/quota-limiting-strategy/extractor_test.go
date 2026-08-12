package quota_limiting_strategy

import (
	"testing"

	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

const LabelParentTenant = "parent_tenant"

type dummyContext struct {
	eocontext.EoContext
	labels map[string]string
}

func (d *dummyContext) GetLabel(key string) string {
	if d.labels == nil {
		return ""
	}
	return d.labels[key]
}

func (d *dummyContext) SetLabel(key string, value string) {
	if d.labels == nil {
		d.labels = make(map[string]string)
	}
	d.labels[key] = value
}

func TestExtractorCRUDAndMultiDimensionDFSMatch(t *testing.T) {
	customerVar = &dummyCustomerVar{
		parents: map[string]string{
			"child_tenant_1": "parent_tenant_1",
		},
	}
	extractor := NewExtractor()

	cfg1 := &Config{
		Na: "strategy_api_1",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 100},
		},
		Filters: FiltersConfig{
			Tenant: "tenant_a",
			Resource: Filter{
				Type:  "api",
				Items: []string{"api_item_123"},
			},
			Target: Filter{
				Type:  "user",
				Items: []string{"user_888"},
			},
		},
	}

	// 1. 测试增加 (Add)
	extractor.AddStrategy(cfg1.Na, cfg1)

	gotCfg, ok := extractor.GetStrategy("strategy_api_1")
	if !ok || gotCfg.Na != "strategy_api_1" {
		t.Fatalf("expected strategy_api_1 to be added, got ok=%v", ok)
	}

	// 2. 测试 5 维多维树 DFS 一步精准匹配成功
	ctxMatch := &dummyContext{
		labels: map[string]string{
			"tenant":                  "tenant_a",
			"resource_type":           "api",
			"resource":                "api_item_123",
			"consumer":                "user_888",
			context_label.LabelConsumerType: "user",
		},
	}

	strategies, has := extractor.GetStrategies(ctxMatch)
	if !has || len(strategies) == 0 {
		t.Fatalf("expected strategy to be matched via full 5-dimension tree DFS")
	}
	if strategies[0].Threshold() != 100 {
		t.Fatalf("expected threshold 100, got %d", strategies[0].Threshold())
	}

	// 3. 测试维不匹配的情况 (DFS 深度切断)
	ctxMismatch := &dummyContext{
		labels: map[string]string{
			"tenant":        "tenant_a",
			"resource_type": "api",
			"resource":      "api_item_999", // 不匹配
			"consumer":      "user_888",
		},
	}

	_, has = extractor.GetStrategies(ctxMismatch)
	if has {
		t.Fatalf("expected strategy NOT to match when resource item differs")
	}

	// 4. 测试更新 (Update)
	cfg1Updated := &Config{
		Na: "strategy_api_1",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 200},
		},
		Filters: FiltersConfig{
			Tenant: "tenant_a",
			Resource: Filter{
				Type:  "api",
				Items: []string{"api_item_123"},
			},
			Target: Filter{
				Type:  "user",
				Items: []string{"user_888"},
			},
		},
	}
	extractor.AddStrategy("strategy_api_1", cfg1Updated)

	strategies, has = extractor.GetStrategies(ctxMatch)
	if !has || len(strategies) == 0 {
		t.Fatalf("expected updated strategy")
	}
	if strategies[0].Threshold() != 200 {
		t.Fatalf("expected threshold 200, got %d", strategies[0].Threshold())
	}

	// 5. 测试删除 (Remove)
	extractor.RemoveStrategy("strategy_api_1")
	_, ok = extractor.GetStrategy("strategy_api_1")
	if ok {
		t.Fatalf("expected strategy_api_1 to be removed")
	}

	_, has = extractor.GetStrategies(ctxMatch)
	if has {
		t.Fatalf("expected no strategies to match after removal")
	}
}

func TestGetParentStrategies(t *testing.T) {
	extractor := NewExtractor()

	// 1. Target 为 All 的父租户策略
	cfgParentAll := &Config{
		Na: "strategy_parent_all",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 300},
		},
		Filters: FiltersConfig{
			Tenant: "parent_tenant_1",
			Resource: Filter{
				All: true,
			},
			Target: Filter{
				Type: "all",
				All:  true,
			},
		},
	}

	// 2. Target 为特定类型的策略（非 All）
	cfgParentSpecific := &Config{
		Na: "strategy_parent_specific",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 500},
		},
		Filters: FiltersConfig{
			Tenant: "parent_tenant_1",
			Resource: Filter{
				All: true,
			},
			Target: Filter{
				Type:  "user",
				Items: []string{"user_123"},
			},
		},
	}

	extractor.AddStrategy(cfgParentAll.Na, cfgParentAll)
	extractor.AddStrategy(cfgParentSpecific.Na, cfgParentSpecific)

	ctx := &dummyContext{
		labels: map[string]string{
			LabelParentTenant: "parent_tenant_1",
			"tenant":          "child_tenant_1",
		},
	}

	strategies, has := extractor.GetParentStrategies(ctx)
	if !has || len(strategies) != 1 {
		t.Fatalf("expected 1 parent strategy with target=all, got has=%v, len=%d", has, len(strategies))
	}
	if strategies[0].Threshold() != 300 {
		t.Fatalf("expected threshold 300 for parent strategy, got %d", strategies[0].Threshold())
	}

	// 测试未指定 parentTenant 的情况
	ctxNoParent := &dummyContext{
		labels: map[string]string{
			"tenant": "tenant_without_parent",
		},
	}
	_, has = extractor.GetParentStrategies(ctxNoParent)
	if has {
		t.Fatalf("expected GetParentStrategies to return false when no parent tenant label")
	}
}

type dummyCustomerVar struct {
	eosc.ICustomerVar
	parents map[string]string
}

func (d *dummyCustomerVar) GetAll(key string) (map[string]string, bool) {
	if len(key) > 7 && key[:7] == "parent:" {
		child := key[7:]
		if parent, ok := d.parents[child]; ok {
			return map[string]string{parent: ""}, true
		}
	}
	return nil, false
}

func TestGetParentStrategiesRecursive(t *testing.T) {
	cv := &dummyCustomerVar{
		parents: map[string]string{
			"curr_tenant": "parent_1",
			"parent_1":    "parent_2",
			"parent_2":    "parent_3",
		},
	}

	// 设置测试专用的 customerVar
	customerVar = cv
	extractor := NewExtractor()

	// 添加 parent_1, parent_2, parent_3 的策略 (Target 都为 All)
	cfgP1 := &Config{
		Na: "strat_p1",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 100},
		},
		Filters: FiltersConfig{
			Tenant:   "parent_1",
			Resource: Filter{All: true},
			Target: Filter{Type: "all", All: true},
		},
	}
	cfgP2 := &Config{
		Na: "strat_p2",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 200},
		},
		Filters: FiltersConfig{
			Tenant:   "parent_2",
			Resource: Filter{All: true},
			Target:   Filter{Type: "all", All: true},
		},
	}
	cfgP3 := &Config{
		Na: "strat_p3",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 300},
		},
		Filters: FiltersConfig{
			Tenant:   "parent_3",
			Resource: Filter{All: true},
			Target:   Filter{Type: "all", All: true},
		},
	}

	extractor.AddStrategy(cfgP1.Na, cfgP1)
	extractor.AddStrategy(cfgP2.Na, cfgP2)
	extractor.AddStrategy(cfgP3.Na, cfgP3)

	ctx := &dummyContext{
		labels: map[string]string{
			"tenant": "curr_tenant",
		},
	}

	strategies, has := extractor.GetParentStrategies(ctx)
	if !has {
		t.Fatalf("expected to find recursive parent strategies")
	}
	if len(strategies) != 3 {
		t.Fatalf("expected 3 recursive parent strategies, got %d", len(strategies))
	}

	thresholds := []int64{strategies[0].Threshold(), strategies[1].Threshold(), strategies[2].Threshold()}
	if thresholds[0] != 300 || thresholds[1] != 200 || thresholds[2] != 100 {
		t.Fatalf("expected thresholds [300, 200, 100] (from top root parent to direct parent), got %v", thresholds)
	}
}

func TestGetExtractorByQuotaType(t *testing.T) {
	extractor := NewExtractor()

	cfg := &Config{
		Na: "strat_multi_quota",
		Quota: QuotaConfig{
			Request:    QuotaRule{Second: 10},
			TotalToken: QuotaRule{Second: 1000},
			Amount:     QuotaRule{Second: 50},
		},
		Filters: FiltersConfig{
			Tenant:   "tenant_x",
			Resource: Filter{All: true},
			Target:   Filter{Type: "all", All: true},
		},
	}

	extractor.AddStrategy(cfg.Na, cfg)

	ctx := &dummyContext{
		labels: map[string]string{
			"tenant": "tenant_x",
		},
	}

	// 1. 通过各种 quotaType 深入查询
	reqStrategies, has := extractor.GetStrategies(ctx, quotaTypeRequest)
	if !has || len(reqStrategies) != 1 || reqStrategies[0].Threshold() != 10 {
		t.Fatalf("expected request threshold 10, got has=%v, len=%d", has, len(reqStrategies))
	}

	tokStrategies, has := extractor.GetStrategies(ctx, quotaTypeTotalToken)
	if !has || len(tokStrategies) != 1 || tokStrategies[0].Threshold() != 1000 {
		t.Fatalf("expected total_token threshold 1000, got has=%v, len=%d", has, len(tokStrategies))
	}

	amtStrategies, has := extractor.GetStrategies(ctx, quotaTypeAmount)
	if !has || len(amtStrategies) != 1 || amtStrategies[0].Threshold() != 50 {
		t.Fatalf("expected amount threshold 50, got has=%v, len=%d", has, len(amtStrategies))
	}

	// 2. 通过参数传递 quotaType 查询
	reqStrategies2, has2 := extractor.GetStrategies(ctx, quotaTypeRequest)
	if !has2 || reqStrategies2[0].Threshold() != 10 {
		t.Fatalf("expected request threshold 10 via parameter")
	}

	// 3. 不带 quotaType 查询，默认走 request 策略
	allStrategies, hasAll := extractor.GetStrategies(ctx)
	if !hasAll || len(allStrategies) != 1 {
		t.Fatalf("expected 1 strategy (request by default) when no quotaType passed, got %d", len(allStrategies))
	}
}
