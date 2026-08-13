package quota_limiting_strategy

import (
	"fmt"
	"testing"

	context_label "github.com/eolinker/apinto/common/context-label"
)

// helper: 快速构建指定数量的大量复杂维度策略
func setupMassiveExtractor(count int) (Extractor, []eocontextLabels) {
	extractor := NewExtractor()
	labelsList := make([]eocontextLabels, 0, count)

	tenantsCount := count / 10
	if tenantsCount < 1 {
		tenantsCount = 1
	}

	for i := 0; i < count; i++ {
		tenant := fmt.Sprintf("tenant_%d", i%tenantsCount)
		apiItem := fmt.Sprintf("api_%d", i%50)
		userItem := fmt.Sprintf("user_%d", i%50)

		cfg := &Config{
			Na: fmt.Sprintf("strat_mass_%d", i),
			Quota: QuotaConfig{
				Request: QuotaRule{Second: int64(100 + i%1000)},
			},
			Filters: FiltersConfig{
				Tenant: tenant,
				Resource: Filter{
					Type:  "api",
					Items: []string{apiItem},
				},
				Target: Filter{
					Type:  "user",
					Items: []string{userItem},
				},
			},
		}
		extractor.AddStrategy(cfg.Na, cfg)

		if i%10 == 0 {
			labelsList = append(labelsList, eocontextLabels{
				tenant:       tenant,
				resourceType: "api",
				resource:     apiItem,
				consumer:     userItem,
				consumerType: "user",
			})
		}
	}

	return extractor, labelsList
}

type eocontextLabels struct {
	tenant       string
	resourceType string
	resource     string
	consumer     string
	consumerType string
}

func (l eocontextLabels) toContext() *dummyContext {
	return &dummyContext{
		labels: map[string]string{
			"tenant":                        l.tenant,
			"resource_type":                 l.resourceType,
			"resource":                      l.resource,
			"consumer":                      l.consumer,
			context_label.LabelConsumerType: l.consumerType,
		},
	}
}

// ---------------- Benchmark 不同策略规模 (Scale 100 / 1000 / 10000) ----------------

func BenchmarkGetStrategies_Scale_100(b *testing.B) {
	extractor, sampleLabels := setupMassiveExtractor(100)
	ctx := sampleLabels[0].toContext()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		extractor.GetStrategies(ctx)
	}
}

func BenchmarkGetStrategies_Scale_1000(b *testing.B) {
	extractor, sampleLabels := setupMassiveExtractor(1000)
	ctx := sampleLabels[0].toContext()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		extractor.GetStrategies(ctx)
	}
}

func BenchmarkGetStrategies_Scale_10000(b *testing.B) {
	extractor, sampleLabels := setupMassiveExtractor(10000)
	ctx := sampleLabels[0].toContext()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		extractor.GetStrategies(ctx)
	}
}

// ---------------- 海量策略 (10,000条) 高并发读压测 ----------------

func BenchmarkGetStrategies_Parallel_10k(b *testing.B) {
	extractor, sampleLabels := setupMassiveExtractor(10000)
	labelsLen := len(sampleLabels)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			ctx := sampleLabels[idx%labelsLen].toContext()
			extractor.GetStrategies(ctx)
			idx++
		}
	})
}

// ---------------- 海量策略 (10,000条) 连续建树/新增压测 ----------------

func BenchmarkAddStrategy_10k(b *testing.B) {
	cfg := &Config{
		Na: "bench_add_strat",
		Quota: QuotaConfig{
			Request: QuotaRule{Second: 1000, Minute: 60000},
		},
		Filters: FiltersConfig{
			Tenant: "bench_tenant",
			Resource: Filter{
				Type:  "api",
				Items: []string{"api_1", "api_2", "api_3"},
			},
			Target: Filter{
				Type:  "user",
				Items: []string{"user_1", "user_2"},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		extractor := NewExtractor()
		for j := 0; j < 100; j++ {
			id := fmt.Sprintf("strat_%d", j)
			extractor.AddStrategy(id, cfg)
		}
	}
}

// ---------------- 多级继承父租户链 (包含海量根/父租户策略) ----------------

func BenchmarkGetStrategies_WithParentChain_Massive(b *testing.B) {
	cv := &dummyCustomerVar{
		parents: map[string]string{
			"child_tenant":   "parent_level_1",
			"parent_level_1": "parent_level_2",
			"parent_level_2": "parent_root",
		},
	}
	customerVar = cv

	extractor := NewExtractor()

	// 根与各父租户下建立大量继承策略
	tenants := []string{"child_tenant", "parent_level_1", "parent_level_2", "parent_root"}
	for _, tName := range tenants {
		for i := 0; i < 250; i++ { // 总共 1000 条父子策略
			cfg := &Config{
				Na: fmt.Sprintf("strat_%s_%d", tName, i),
				Quota: QuotaConfig{
					Request: QuotaRule{Second: int64(1000 + i)},
				},
				Filters: FiltersConfig{
					Tenant:   tName,
					Resource: Filter{All: true},
					Target:   Filter{Type: "all", All: true},
				},
			}
			extractor.AddStrategy(cfg.Na, cfg)
		}
	}

	ctx := &dummyContext{
		labels: map[string]string{
			"tenant": "child_tenant",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		extractor.GetParentStrategies(ctx)
	}
}

// ---------------- 底层 6 维树在 10,000 个路径分支下的 DFS 检索压测 ----------------

func BenchmarkGenericDimensionTree_Search_10k(b *testing.B) {
	tree := NewGenericDimensionTree[string]()

	// 插入 10,000 条维度路径分支
	for i := 0; i < 10000; i++ {
		path := []string{
			fmt.Sprintf("tenant_%d", i%100),
			"user",
			fmt.Sprintf("user_%d", i%50),
			"api",
			"all",
			fmt.Sprintf("res_%d", i%50),
		}
		tree.Insert(path, fmt.Sprintf("item_%d", i))
	}

	queryKeys := [][]string{
		{"tenant_25"},
		{"all", "user"},
		{"all", "user_10"},
		{"all", "api"},
		{"all"},
		{"all", "res_15"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tree.Search(queryKeys)
	}
}
