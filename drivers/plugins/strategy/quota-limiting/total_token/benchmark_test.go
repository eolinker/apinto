package total_token

import (
	"context"
	"fmt"
	"testing"
	"time"

	context_label "github.com/eolinker/apinto/common/context-label"
	quota_limiting_strategy "github.com/eolinker/apinto/drivers/strategy/quota-limiting-strategy"
	price_calcular "github.com/eolinker/apinto/price-calcular"
)

// BenchmarkBuildQuotaKeyAndTTL 测试不同周期下 Key 生成与 TTL 计算开销
func BenchmarkBuildQuotaKeyAndTTL(b *testing.B) {
	ctx := &dummyHttpContext{
		labels: map[string]string{
			"product":     "apinto",
			"application": "app_bench_1",
		},
	}
	st := &Strategy{
		key: context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"),
	}
	now := time.Now()

	periods := []quota_limiting_strategy.Period{
		quota_limiting_strategy.PeriodSecond,
		quota_limiting_strategy.PeriodMinute,
		quota_limiting_strategy.PeriodHour,
		quota_limiting_strategy.PeriodDay,
		quota_limiting_strategy.PeriodMonth,
		quota_limiting_strategy.PeriodTotal,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		p := periods[i%len(periods)]
		dummySt := &dummyStrategy{
			id:         "strat_bench",
			targetType: "user",
			period:     p,
			threshold:  100000,
		}
		st.buildQuotaKeyAndTTL(ctx, dummySt, now)
	}
}

// BenchmarkDoHttpFilter_SingleStrategy 单策略 DoHttpFilter 完整过滤流程基准测试
func BenchmarkDoHttpFilter_SingleStrategy(b *testing.B) {
	calcId := "api:res_bench_single"
	calc := createTokenCalculator(calcId)
	price_calcular.SetCalculator(calcId, calc)
	defer price_calcular.DelCalculator(calcId)

	cfg := &quota_limiting_strategy.Config{
		Na: "strat_bench_single",
		Quota: quota_limiting_strategy.QuotaConfig{
			TotalToken: quota_limiting_strategy.QuotaRule{Hour: 100000000},
		},
		Filters: quota_limiting_strategy.FiltersConfig{
			Tenant:   "tenant_bench_single",
			Resource: quota_limiting_strategy.Filter{All: true},
			Target:   quota_limiting_strategy.Filter{Type: "all", All: true},
		},
	}
	quota_limiting_strategy.Create("worker_bench_single", "strat_bench_single", cfg, nil)

	st := &Strategy{
		key: context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"),
	}

	chain := &dummyChain{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := newDummyHttpContext()
		ctx.SetLabel("api", "api_bench")
		ctx.SetLabel("resource_type", "api")
		ctx.SetLabel("resource", "res_bench_single")
		ctx.SetLabel("tenant", "tenant_bench_single")
		context_label.SetPreInputToken(ctx, 100)

		_ = st.DoHttpFilter(ctx, chain)
	}
}

// BenchmarkDoHttpFilter_Parallel 网关极高并发场景下的基准压测
func BenchmarkDoHttpFilter_Parallel(b *testing.B) {
	calcId := "api:res_bench_parallel"
	calc := createTokenCalculator(calcId)
	price_calcular.SetCalculator(calcId, calc)
	defer price_calcular.DelCalculator(calcId)

	cfg := &quota_limiting_strategy.Config{
		Na: "strat_bench_parallel",
		Quota: quota_limiting_strategy.QuotaConfig{
			TotalToken: quota_limiting_strategy.QuotaRule{Hour: 1000000000},
		},
		Filters: quota_limiting_strategy.FiltersConfig{
			Tenant:   "tenant_bench_parallel",
			Resource: quota_limiting_strategy.Filter{All: true},
			Target:   quota_limiting_strategy.Filter{Type: "all", All: true},
		},
	}
	quota_limiting_strategy.Create("worker_bench_parallel", "strat_bench_parallel", cfg, nil)

	st := &Strategy{
		key: context_label.NewKeyGenerator("{product}:quota-limiting:{strategy}:{target_type}:{application}:{period}:{time_format}"),
	}

	chain := &dummyChain{}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			idx++
			ctx := newDummyHttpContext()
			ctx.SetLabel("api", fmt.Sprintf("api_bench_%d", idx%10))
			ctx.SetLabel("resource_type", "api")
			ctx.SetLabel("resource", "res_bench_parallel")
			ctx.SetLabel("tenant", "tenant_bench_parallel")
			context_label.SetPreInputToken(ctx, 100)

			_ = st.DoHttpFilter(ctx, chain)
		}
	})
}

// BenchmarkSettle 压测补扣/多退少补结算阶段耗时
func BenchmarkSettle(b *testing.B) {
	mc := newMockCache()
	ctx := newDummyHttpContext()
	items := []*executedItem{
		{key: "key_1", ttl: time.Hour},
		{key: "key_2", ttl: time.Hour},
	}

	mc.IncrBy(context.Background(), "key_1", 1000, time.Hour)
	mc.IncrBy(context.Background(), "key_2", 1000, time.Hour)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		context_label.SetActualTotalToken(ctx, 1200) // 实际比预扣多 200，触发补扣
		settle(ctx, mc, items, 1000)
	}
}
