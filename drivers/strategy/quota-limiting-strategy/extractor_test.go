package quota_limiting_strategy

import (
	"fmt"
	"sync"
	"testing"

	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/apinto/utils/response"
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

// 1. Config & Check 单元测试
func TestConfigCheck(t *testing.T) {
	cfg := &Config{
		Na:   "test_name",
		Desc: "test_desc",
	}

	if cfg.Name() != "test_name" {
		t.Errorf("expected test_name, got %s", cfg.Name())
	}
	if cfg.Description() != "test_desc" {
		t.Errorf("expected test_desc, got %s", cfg.Description())
	}
	if err := cfg.Check(); err != nil {
		t.Errorf("expected nil error for valid config, got %v", err)
	}

	if err := checkConfig(nil); err == nil {
		t.Error("expected error when checking nil config, got nil")
	}
}

// 2. Period 与 Strategy 排序逻辑测试
func TestPeriodAndStrategySorting(t *testing.T) {
	// 2.1 测试 Period.String()
	periods := map[Period]string{
		PeriodSecond: "second",
		PeriodMinute: "minute",
		PeriodHour:   "hour",
		PeriodDay:    "day",
		PeriodMonth:  "month",
		PeriodTotal:  "total",
		Period(99):   "second", // default
	}

	for p, expected := range periods {
		if p.String() != expected {
			t.Errorf("Period(%d).String() = %s, expected %s", p, p.String(), expected)
		}
	}

	// 2.2 NewStrategies 生成 6 维策略
	rule := QuotaRule{
		Second: 10,
		Minute: 100,
		Hour:   1000,
		Day:    10000,
		Month:  100000,
		Total:  1000000,
	}
	resp := response.Parse(nil)
	strats := NewStrategies("st_1", "user", rule, resp)
	if len(strats) != 6 {
		t.Fatalf("expected 6 strategies, got %d", len(strats))
	}

	// 2.3 sortStrategies 排序测试
	unsorted := []IStrategy{
		&Strategy{period: PeriodHour, threshold: 500},
		&Strategy{period: PeriodSecond, threshold: 20},
		&Strategy{period: PeriodSecond, threshold: 10},
		&Strategy{period: PeriodMinute, threshold: 100},
	}

	sortStrategies(unsorted)

	if unsorted[0].Period() != PeriodSecond || unsorted[0].Threshold() != 10 {
		t.Errorf("expected 1st strategy PeriodSecond threshold 10, got period=%v, threshold=%d", unsorted[0].Period(), unsorted[0].Threshold())
	}
	if unsorted[1].Period() != PeriodSecond || unsorted[1].Threshold() != 20 {
		t.Errorf("expected 2nd strategy PeriodSecond threshold 20, got period=%v, threshold=%d", unsorted[1].Period(), unsorted[1].Threshold())
	}
	if unsorted[2].Period() != PeriodMinute || unsorted[2].Threshold() != 100 {
		t.Errorf("expected 3rd strategy PeriodMinute threshold 100, got period=%v, threshold=%d", unsorted[2].Period(), unsorted[2].Threshold())
	}
	if unsorted[3].Period() != PeriodHour || unsorted[3].Threshold() != 500 {
		t.Errorf("expected 4th strategy PeriodHour threshold 500, got period=%v, threshold=%d", unsorted[3].Period(), unsorted[3].Threshold())
	}
}

// 3. Extractor 基本增删改查与 5 维多树匹配测试
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
			"tenant":                        "tenant_a",
			"resource_type":                 "api",
			"resource":                      "api_item_123",
			"consumer":                      "user_888",
			context_label.LabelConsumerType: "user",
		},
	}

	tenantStrats, has := extractor.GetStrategies(ctxMatch)
	if !has || len(tenantStrats) == 0 {
		t.Fatalf("expected strategy to be matched via full 5-dimension tree DFS")
	}
	strats := tenantStrats[0].Strategies()
	if len(strats) == 0 || strats[0].Threshold() != 100 {
		t.Fatalf("expected threshold 100, got %v", strats)
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

	tenantStrats, has = extractor.GetStrategies(ctxMatch)
	if !has || len(tenantStrats) == 0 {
		t.Fatalf("expected updated strategy")
	}
	strats = tenantStrats[0].Strategies()
	if len(strats) == 0 || strats[0].Threshold() != 200 {
		t.Fatalf("expected threshold 200, got %v", strats)
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

// 4. Channel 模式 Target 匹配测试
func TestChannelTargetFilter(t *testing.T) {
	extractor := NewExtractor()

	cfgChannel := &Config{
		Na: "strategy_channel",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 50},
		},
		Filters: FiltersConfig{
			Tenant: "tenant_channel",
			Resource: Filter{
				Type: "api",
				All:  true,
			},
			Target: Filter{
				Type:  "channel",
				Items: []string{"channel_app_1", "channel_app_2"},
			},
		},
	}

	extractor.AddStrategy(cfgChannel.Na, cfgChannel)

	// 测试匹配 channel_app_1
	ctx1 := &dummyContext{
		labels: map[string]string{
			"tenant":        "channel_app_1", // Target.Type == "channel" 时，tenant 节点为 channel.Items 中的元素
			"resource_type": "api",
		},
	}
	tenantStrats, has := extractor.GetStrategies(ctx1)
	if !has || len(tenantStrats) == 0 {
		t.Fatalf("expected channel_app_1 strategy matched")
	}

	// 测试未包含的 channel_app_3
	ctx3 := &dummyContext{
		labels: map[string]string{
			"tenant": "channel_app_3",
		},
	}
	_, has = extractor.GetStrategies(ctx3)
	if has {
		t.Fatalf("expected channel_app_3 NOT matched")
	}

	// 测试 channel Items 为空的情况（无法生成路径）
	cfgChannelEmpty := &Config{
		Na: "strategy_channel_empty",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 50},
		},
		Filters: FiltersConfig{
			Tenant: "tenant_channel",
			Target: Filter{
				Type:  "channel",
				Items: []string{}, // 空
			},
		},
	}
	extractor.AddStrategy(cfgChannelEmpty.Na, cfgChannelEmpty)
	_, ok := extractor.GetStrategy(cfgChannelEmpty.Na)
	if ok {
		t.Fatalf("strategy with empty channel items should not be indexed")
	}
}

// 5. 无效 Tenant 及边界测试
func TestInvalidTenantAndBoundary(t *testing.T) {
	extractor := NewExtractor()

	// 5.1 无效 tenant (空, "all", "*")
	invalidTenants := []string{"", "all", "*"}
	for i, tenant := range invalidTenants {
		cfg := &Config{
			Na: fmt.Sprintf("invalid_tenant_%d", i),
			Quota: QuotaConfig{
				Request: QuotaRule{Second: 10},
			},
			Filters: FiltersConfig{
				Tenant: tenant,
				Target: Filter{Type: "all", All: true},
			},
		}
		extractor.AddStrategy(cfg.Na, cfg)
		if _, ok := extractor.GetStrategy(cfg.Na); ok {
			t.Errorf("strategy with invalid tenant %q should not be added", tenant)
		}
	}

	// 5.2 AddStrategy nil/空配置防护
	extractor.AddStrategy("", nil)
	extractor.AddStrategy("some_id", nil)

	// 5.3 RemoveStrategy 空 ID 防护
	extractor.RemoveStrategy("")

	// 5.4 GetStrategies 传入空 Context 或无 Tenant
	ctxNoTenant := &dummyContext{labels: map[string]string{}}
	if _, has := extractor.GetStrategies(ctxNoTenant); has {
		t.Error("expected GetStrategies to return false when context has no tenant")
	}
}

// 6. 父租户策略获取测试
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

	extractor.AddStrategy(cfgParentAll.Na, cfgParentAll)

	ctx := &dummyContext{
		labels: map[string]string{
			LabelParentTenant: "parent_tenant_1",
			"tenant":          "child_tenant_1",
		},
	}

	tenantStrats, has := extractor.GetParentStrategies(ctx)
	if !has || len(tenantStrats) != 1 {
		t.Fatalf("expected 1 parent strategy with target=all, got has=%v, len=%d", has, len(tenantStrats))
	}
	strats := tenantStrats[0].Strategies()
	if len(strats) == 0 || strats[0].Threshold() != 300 {
		t.Fatalf("expected threshold 300 for parent strategy, got %v", strats)
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

// 7. 递归向上获取多层父租户策略
func TestGetParentStrategiesRecursive(t *testing.T) {
	cv := &dummyCustomerVar{
		parents: map[string]string{
			"curr_tenant": "parent_1",
			"parent_1":    "parent_2",
			"parent_2":    "parent_3",
		},
	}

	customerVar = cv
	extractor := NewExtractor()

	cfgP1 := &Config{
		Na: "strat_p1",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 100},
		},
		Filters: FiltersConfig{
			Tenant:   "parent_1",
			Resource: Filter{All: true},
			Target:   Filter{Type: "all", All: true},
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

	tenantStrats, has := extractor.GetParentStrategies(ctx)
	if !has {
		t.Fatalf("expected to find recursive parent strategies")
	}

	// 所有属于该租户链的策略会按层次追加合并
	var thresholds []int64
	for _, ts := range tenantStrats {
		for _, s := range ts.Strategies() {
			thresholds = append(thresholds, s.Threshold())
		}
	}

	if len(thresholds) != 3 {
		t.Fatalf("expected 3 recursive parent strategies, got %d", len(thresholds))
	}

	// 从根租户 parent_3(300) -> parent_2(200) -> parent_1(100)
	if thresholds[0] != 300 || thresholds[1] != 200 || thresholds[2] != 100 {
		t.Fatalf("expected thresholds [300, 200, 100] (from top root parent to direct parent), got %v", thresholds)
	}
}

// 8. 多额度类型 (Request, TotalToken, Amount) 关联查询测试及全局函数
func TestGetExtractorByQuotaTypeAndGlobalHelpers(t *testing.T) {
	extractorManager = NewExtractor()

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

	extractorManager.AddStrategy(cfg.Na, cfg)

	ctx := &dummyContext{
		labels: map[string]string{
			"tenant": "tenant_x",
		},
	}

	// 1. 通过全局帮助函数验证
	reqStrats, hasReq := GetRequestStrategies(ctx)
	if !hasReq || reqStrats[0].Strategies()[0].Threshold() != 10 {
		t.Fatalf("expected request threshold 10, got has=%v", hasReq)
	}

	tokStrats, hasTok := GetTotalTokenStrategies(ctx)
	if !hasTok || tokStrats[0].Strategies()[0].Threshold() != 1000 {
		t.Fatalf("expected total_token threshold 1000, got has=%v", hasTok)
	}

	amtStrats, hasAmt := GetAmountStrategies(ctx)
	if !hasAmt || amtStrats[0].Strategies()[0].Threshold() != 50 {
		t.Fatalf("expected amount threshold 50, got has=%v", hasAmt)
	}

	// 清理
	removeStrategy(cfg.Na)
}

// 9. 并发读写测试
func TestConcurrentAccess(t *testing.T) {
	extractor := NewExtractor()

	var wg sync.WaitGroup
	workers := 10
	iterations := 50

	// 并发写
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				id := fmt.Sprintf("strat_%d_%d", workerID, j)
				cfg := &Config{
					Na: id,
					Quota: QuotaConfig{
						Request: QuotaRule{Second: int64(j + 1)},
					},
					Filters: FiltersConfig{
						Tenant:   fmt.Sprintf("tenant_%d", workerID),
						Resource: Filter{All: true},
						Target:   Filter{Type: "all", All: true},
					},
				}
				extractor.AddStrategy(id, cfg)
				extractor.GetStrategy(id)

				ctx := &dummyContext{
					labels: map[string]string{
						"tenant": fmt.Sprintf("tenant_%d", workerID),
					},
				}
				extractor.GetStrategies(ctx)

				if j%2 == 0 {
					extractor.RemoveStrategy(id)
				}
			}
		}(i)
	}

	wg.Wait()
}
