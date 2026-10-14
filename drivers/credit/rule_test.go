package credit

import (
	"fmt"
	"sync"
	"testing"
)

func TestMatchCandidateRules(t *testing.T) {
	mgr := NewManager()

	r1 := CompileRule("rule_1", &Config{
		Filters: Filters{
			Provider: RuleFilter{
				AI: []string{"openai"},
			},
			Resource: RuleFilter{
				AI: []string{"mod_gpt4"},
			},
		},
		Nbf: 1000,
		Exp: 3000,
	})

	r2 := CompileRule("rule_2", &Config{
		Filters: Filters{
			Provider: RuleFilter{
				AI: []string{"all"},
			},
			Resource: RuleFilter{
				AI: []string{"mod_gpt4"},
			},
		},
		Nbf: 1000,
		Exp: 2000, // 过期更早，应该排在前面
	})

	r3 := CompileRule("rule_3", &Config{
		Filters: Filters{
			Provider: RuleFilter{
				API: []string{"serv_123"},
			},
			Resource: RuleFilter{
				API: []string{"api_456"},
			},
		},
		Nbf: 1000,
		Exp: 5000,
	})

	mgr.SetRule("rule_1", r1)
	mgr.SetRule("rule_2", r2)
	mgr.SetRule("rule_3", r3)

	now := int64(1500)

	// 测试 AI 匹配
	candidates := mgr.MatchCandidateRules("ai", "openai", "mod_gpt4", now)
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates for AI, got %d", len(candidates))
	}
	// Exp 排序测试: rule_2 (exp 2000) 必须先于 rule_1 (exp 3000)
	if candidates[0].ID != "rule_2" || candidates[1].ID != "rule_1" {
		t.Fatalf("expected rule_2 then rule_1, got %s then %s", candidates[0].ID, candidates[1].ID)
	}

	// 测试 API 匹配
	apiCandidates := mgr.MatchCandidateRules("api", "serv_123", "api_456", now)
	if len(apiCandidates) != 1 || apiCandidates[0].ID != "rule_3" {
		t.Fatalf("expected rule_3 for API, got %v", apiCandidates)
	}

	// 测试过期
	expiredCandidates := mgr.MatchCandidateRules("ai", "openai", "mod_gpt4", 2500)
	if len(expiredCandidates) != 1 || expiredCandidates[0].ID != "rule_1" {
		t.Fatalf("expected rule_1 (rule_2 expired at 2000), got %v", expiredCandidates)
	}
}

func TestRuleTree_Combinations(t *testing.T) {
	mgr := NewManager()

	// r1: provider=p1, resource=res_a
	mgr.SetRule("r1", CompileRule("r1", &Config{
		Filters: Filters{
			Provider: RuleFilter{AI: []string{"p1"}},
			Resource: RuleFilter{AI: []string{"res_a"}},
		},
		Exp: 100,
	}))

	// r2: provider=*, resource=res_a
	mgr.SetRule("r2", CompileRule("r2", &Config{
		Filters: Filters{
			Provider: RuleFilter{AI: []string{"*"}},
			Resource: RuleFilter{AI: []string{"res_a"}},
		},
		Exp: 200,
	}))

	// r3: provider=p1, resource=res_b
	mgr.SetRule("r3", CompileRule("r3", &Config{
		Filters: Filters{
			Provider: RuleFilter{AI: []string{"p1"}},
			Resource: RuleFilter{AI: []string{"res_b"}},
		},
		Exp: 300,
	}))

	// 查 (p1, res_a) -> 应该命中 (r1, r2) 按 exp 排序
	res := mgr.MatchCandidateRules("ai", "p1", "res_a", 50)
	if len(res) != 2 || res[0].ID != "r1" || res[1].ID != "r2" {
		t.Fatalf("expected [r1, r2], got %v", res)
	}

	// 查 (p2, res_a) -> 应该命中 provider=* 的 r2
	res = mgr.MatchCandidateRules("ai", "p2", "res_a", 50)
	if len(res) != 1 || res[0].ID != "r2" {
		t.Fatalf("expected [r2], got %v", res)
	}

	// 查 (p1, res_b) -> 应该命中 r3
	res = mgr.MatchCandidateRules("ai", "p1", "res_b", 50)
	if len(res) != 1 || res[0].ID != "r3" {
		t.Fatalf("expected [r3], got %v", res)
	}

	// 查 (p2, res_b) -> 无通配 provider 规则，应该为 0
	res = mgr.MatchCandidateRules("ai", "p2", "res_b", 50)
	if len(res) != 0 {
		t.Fatalf("expected 0 rules, got %d", len(res))
	}

	// 查未配置的具体 resource -> 0（resource 不走通配）
	res = mgr.MatchCandidateRules("ai", "p1", "res_c", 50)
	if len(res) != 0 {
		t.Fatalf("expected 0 rules for res_c, got %d", len(res))
	}

	// 查未支持的协议类型 -> 0
	res = mgr.MatchCandidateRules("api", "p1", "res_a", 50)
	if len(res) != 0 {
		t.Fatalf("expected 0 rules for api, got %d", len(res))
	}

	// 测试删除规则
	mgr.DelRule("r2")
	res = mgr.MatchCandidateRules("ai", "p2", "res_a", 50)
	if len(res) != 0 {
		t.Fatalf("expected 0 rules after r2 deleted, got %d", len(res))
	}
}

func TestRuleTree_SortTieBreaker(t *testing.T) {
	mgr := NewManager()

	mgr.SetRule("beta", CompileRule("beta", &Config{
		Filters: Filters{Provider: RuleFilter{AI: []string{"*"}}, Resource: RuleFilter{AI: []string{"res_1"}}},
		Exp:     100,
	}))
	mgr.SetRule("alpha", CompileRule("alpha", &Config{
		Filters: Filters{Provider: RuleFilter{AI: []string{"*"}}, Resource: RuleFilter{AI: []string{"res_1"}}},
		Exp:     100,
	}))

	res := mgr.MatchCandidateRules("ai", "p", "res_1", 0)
	if len(res) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(res))
	}
	// Exp 相同，ID 字典序正序
	if res[0].ID != "alpha" || res[1].ID != "beta" {
		t.Fatalf("expected [alpha, beta], got [%s, %s]", res[0].ID, res[1].ID)
	}
}

func TestRuleTree_Concurrency(t *testing.T) {
	mgr := NewManager()

	for i := 0; i < 20; i++ {
		id := fmt.Sprintf("rule_%d", i)
		mgr.SetRule(id, CompileRule(id, &Config{
			Filters: Filters{
				Provider: RuleFilter{AI: []string{"prov"}},
				Resource: RuleFilter{AI: []string{"res"}},
			},
			Exp: int64(100 + i),
		}))
	}

	var wg sync.WaitGroup
	// 50 个并发读 goroutines
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k := 0; k < 100; k++ {
				_ = mgr.MatchCandidateRules("ai", "prov", "res", 50)
			}
		}()
	}

	// 5 个并发写 goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for k := 0; k < 20; k++ {
				id := fmt.Sprintf("dyn_rule_%d_%d", workerID, k)
				mgr.SetRule(id, CompileRule(id, &Config{
					Filters: Filters{
						Provider: RuleFilter{AI: []string{"prov"}},
						Resource: RuleFilter{AI: []string{"res"}},
					},
					Exp: 500,
				}))
				mgr.DelRule(id)
			}
		}(i)
	}

	wg.Wait()
}

func BenchmarkRuleTree_Match(b *testing.B) {
	mgr := NewManager()

	// 预先注册 200 条不同维度的代金券规则
	for i := 0; i < 200; i++ {
		id := fmt.Sprintf("rule_%d", i)
		p := fmt.Sprintf("provider_%d", i%10)
		r := fmt.Sprintf("resource_%d", i%20)
		if i%5 == 0 {
			p = "*"
		}
		mgr.SetRule(id, CompileRule(id, &Config{
			Filters: Filters{
				Provider: RuleFilter{AI: []string{p}},
				Resource: RuleFilter{AI: []string{r}},
			},
			Exp: int64(1000 + i),
		}))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = mgr.MatchCandidateRules("ai", "provider_2", "resource_4", 500)
		}
	})
}

func TestMatchCandidateRules_MultiDimensionOrder(t *testing.T) {
	mgr := NewManager()

	// rule_prov_only: 只有 provider 没配 resource -> 不应命中
	mgr.SetRule("rule_prov_only", CompileRule("rule_prov_only", &Config{
		Filters: Filters{
			Provider: RuleFilter{AI: []string{"openai"}},
		},
		Exp: 2000,
	}))

	// rule_res_only: 只有 resource 没配 provider (自动通配 provider) -> 可以命中
	mgr.SetRule("rule_res_only", CompileRule("rule_res_only", &Config{
		Filters: Filters{
			Resource: RuleFilter{AI: []string{"gpt-4"}},
		},
		Exp: 3000,
	}))

	// rule_both: 同时指定 provider 和 resource
	mgr.SetRule("rule_both", CompileRule("rule_both", &Config{
		Filters: Filters{
			Provider: RuleFilter{AI: []string{"openai"}},
			Resource: RuleFilter{AI: []string{"gpt-4"}},
		},
		Exp: 1000,
	}))

	// 1. 匹配 openai + gpt-4
	// provider 维度候选: rule_prov_only, rule_res_only (*), rule_both
	// resource 维度候选: rule_res_only, rule_both
	// 交集: rule_res_only, rule_both
	// 排序: rule_both (Exp 1000) 先于 rule_res_only (Exp 3000)
	cands := mgr.MatchCandidateRules("ai", "openai", "gpt-4", 500)
	if len(cands) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(cands))
	}
	if cands[0].ID != "rule_both" || cands[1].ID != "rule_res_only" {
		t.Fatalf("expected [rule_both, rule_res_only], got [%s, %s]", cands[0].ID, cands[1].ID)
	}

	// 2. 匹配 anthropic + gpt-4
	// provider 维度候选: rule_res_only (*)
	// resource 维度候选: rule_res_only, rule_both
	// 交集: rule_res_only
	candsAnthropic := mgr.MatchCandidateRules("ai", "anthropic", "gpt-4", 500)
	if len(candsAnthropic) != 1 || candsAnthropic[0].ID != "rule_res_only" {
		t.Fatalf("expected [rule_res_only], got %v", candsAnthropic)
	}
}
