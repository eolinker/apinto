package extractor

import (
	"testing"

	"github.com/eolinker/eosc/eocontext"
)

type mockItem struct {
	id  string
	val string
}

func (m *mockItem) Name() string {
	if m == nil {
		return ""
	}
	return m.id
}

type mockEoContext struct {
	eocontext.EoContext
	labels map[string]string
}

func (m *mockEoContext) GetLabel(name string) string {
	if m.labels == nil {
		return ""
	}
	return m.labels[name]
}

func TestExtractor_ExactAndWildcard(t *testing.T) {
	ext := NewExtractor[*mockItem]("provider")

	itemExact := &mockItem{id: "openai_rule", val: "openai"}
	itemWildcard := &mockItem{id: "all_provider_rule", val: "*"}

	ext.Set("rule_exact", []string{"openai"}, []*mockItem{itemExact})
	ext.Set("rule_all", []string{"*"}, []*mockItem{itemWildcard})

	// 查 openai: 应该拿到精确的 openai_rule 以及通配的 all_provider_rule
	gotOpenAI, ok := ext.GetByValue("openai")
	if !ok || len(gotOpenAI) != 2 {
		t.Fatalf("expected 2 items for openai (exact + wildcard), got %d (ok=%v)", len(gotOpenAI), ok)
	}

	// 查 unknown: 应该只拿到通配的 all_provider_rule
	gotUnknown, ok := ext.GetByValue("anthropic")
	if !ok || len(gotUnknown) != 1 || gotUnknown[0].id != "all_provider_rule" {
		t.Fatalf("expected wildcard item for unknown, got %v", gotUnknown)
	}

	// 查 "" 或 "*": 应该拿到通配的 all_provider_rule
	gotWildcard, ok := ext.GetByValue("*")
	if !ok || len(gotWildcard) != 1 || gotWildcard[0].id != "all_provider_rule" {
		t.Fatalf("expected wildcard item for '*', got %v", gotWildcard)
	}

	// 测试删除
	ext.Del("rule_all")
	gotAfterDel, ok := ext.GetByValue("anthropic")
	if ok || len(gotAfterDel) != 0 {
		t.Fatalf("expected no items after rule_all deleted, got %v", gotAfterDel)
	}
}

func TestExtractor_ContextAndGlobal(t *testing.T) {
	// 1. 测试命名维度 Get(ctx)
	ext := NewExtractor[*mockItem]("provider")
	item := &mockItem{id: "p1", val: "openai"}
	ext.Set("id1", []string{"openai"}, []*mockItem{item})

	ctx := &mockEoContext{
		labels: map[string]string{
			"provider": "openai",
		},
	}
	res, ok := ext.Get(ctx)
	if !ok || len(res) != 1 || res[0].id != "p1" {
		t.Fatalf("expected [p1], got %v", res)
	}

	// 2. 测试全局通用维度 ("")
	globalExt := NewExtractor[*mockItem]("")
	gItem := &mockItem{id: "g1", val: "global"}
	globalExt.Set("gid1", nil, []*mockItem{gItem})

	gRes, gOk := globalExt.Get(ctx)
	if !gOk || len(gRes) != 1 || gRes[0].id != "g1" {
		t.Fatalf("expected global [g1], got %+v", gRes)
	}

	globalExt.Del("gid1")
	gResAfterDel, gOkAfterDel := globalExt.Get(ctx)
	if gOkAfterDel || len(gResAfterDel) != 0 {
		t.Fatalf("expected empty after del, got %v", gResAfterDel)
	}
}

func TestManager_GetByValuesOrder(t *testing.T) {
	mgr := NewManager[*mockItem]("provider", "resource")

	extProv := NewExtractor[*mockItem]("provider")
	extRes := NewExtractor[*mockItem]("resource")

	pItem := &mockItem{id: "p1", val: "openai"}
	rItem := &mockItem{id: "r1", val: "gpt-4"}

	extProv.Set("p_strat", []string{"openai"}, []*mockItem{pItem})
	extRes.Set("r_strat", []string{"gpt-4"}, []*mockItem{rItem})

	mgr.SetExtractor("provider", extProv)
	mgr.SetExtractor("resource", extRes)

	vals := map[string]string{
		"provider": "openai",
		"resource": "gpt-4",
	}

	// 指定顺序：先 provider，再 resource
	results := mgr.GetByValuesOrder(vals, "provider", "resource")
	if len(results) != 2 {
		t.Fatalf("expected 2 dimension results, got %d", len(results))
	}
	if results[0].Dimension != "provider" || len(results[0].Items) != 1 || results[0].Items[0].id != "p1" {
		t.Errorf("expected 1st dimension to be provider with p1, got %+v", results[0])
	}
	if results[1].Dimension != "resource" || len(results[1].Items) != 1 || results[1].Items[0].id != "r1" {
		t.Errorf("expected 2nd dimension to be resource with r1, got %+v", results[1])
	}

	// 逆序指定：先 resource，再 provider
	reverseResults := mgr.GetByValuesOrder(vals, "resource", "provider")
	if len(reverseResults) != 2 {
		t.Fatalf("expected 2 dimension results, got %d", len(reverseResults))
	}
	if reverseResults[0].Dimension != "resource" || reverseResults[1].Dimension != "provider" {
		t.Errorf("expected order resource, provider, got %+v", reverseResults)
	}

	// 测试 Context 驱动的 GetByOrder
	ctx := &mockEoContext{
		labels: map[string]string{
			"provider": "openai",
			"resource": "gpt-4",
		},
	}
	ctxResults := mgr.GetByOrder(ctx)
	if len(ctxResults) != 2 || ctxResults[0].Dimension != "provider" || ctxResults[1].Dimension != "resource" {
		t.Fatalf("expected default order [provider, resource], got %+v", ctxResults)
	}
}
