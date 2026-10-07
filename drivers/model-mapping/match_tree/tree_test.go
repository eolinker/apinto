package match_tree

import (
	"testing"

	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc/eocontext"
)

type mockContext struct {
	eocontext.EoContext
	labels map[string]string
}

func (m *mockContext) GetLabel(name string) string {
	if m.labels == nil {
		return ""
	}
	return m.labels[name]
}

func (m *mockContext) SetLabel(name, value string) {
	if m.labels == nil {
		m.labels = make(map[string]string)
	}
	m.labels[name] = value
}

func TestGenericDimensionTree(t *testing.T) {
	tree := NewGenericDimensionTree[string]()

	// 插入数据
	tree.Insert([]string{"consumer_1", "claude-opus-5"}, "strat_1")
	tree.Insert([]string{"consumer_1", "gpt-4o"}, "strat_2")
	tree.Insert([]string{"consumer_2", "claude-opus-5"}, "strat_3")

	// 搜索 consumer_1 + claude-opus-5
	results := tree.Search([][]string{{"consumer_1"}, {"claude-opus-5"}})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if len(results[0].Items) != 1 || results[0].Items[0] != "strat_1" {
		t.Fatalf("expected strat_1, got %+v", results[0].Items)
	}

	// 搜索不存在的路径
	results = tree.Search([][]string{{"consumer_2"}, {"gpt-4o"}})
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}

	// 删除
	rem := tree.Remove([]string{"consumer_1", "claude-opus-5"}, func(item string) bool {
		return item == "strat_1"
	})
	if rem != 0 {
		t.Fatalf("expected 0 remaining items, got %d", rem)
	}

	results = tree.Search([][]string{{"consumer_1"}, {"claude-opus-5"}})
	if len(results) != 0 {
		t.Fatalf("expected 0 result after remove, got %d", len(results))
	}
}

func TestLabelMatchTree_NoFallback(t *testing.T) {
	// 定义层级顺序：0: consumer, 1: model
	matchTree := NewLabelMatchTree[string]("consumer", "model")

	// 1. 测试配置缺失任意维度：不生成路径
	filterIncomplete := strategy.FilterConfig{
		"consumer": []string{"c1"},
		// 缺少 model
	}
	paths := matchTree.GenerateDimensionPaths(filterIncomplete)
	if len(paths) != 0 {
		t.Fatalf("expected nil or empty paths when dimension is missing, got %v", paths)
	}

	// 2. 正常配置且包含多候选展开
	filterComplete := strategy.FilterConfig{
		"consumer": []string{"c1"},
		"model":    []string{"claude-opus-5", "claude-opus-pro"},
	}
	paths = matchTree.GenerateDimensionPaths(filterComplete)
	if len(paths) != 2 {
		t.Fatalf("expected 2 cartesian paths, got %d", len(paths))
	}

	for _, p := range paths {
		matchTree.Tree().Insert(p, "strategy_c1")
	}

	// 3. 上下文缺少维度（例如未提取到 consumer 或 model）：立即短路返回 nil
	ctxNoConsumer := &mockContext{labels: map[string]string{"model": "claude-opus-5"}}
	if res := matchTree.Match(ctxNoConsumer); len(res) != 0 {
		t.Fatalf("expected nil when context missing consumer, got %v", res)
	}

	ctxNoModel := &mockContext{labels: map[string]string{"consumer": "c1"}}
	if res := matchTree.Match(ctxNoModel); len(res) != 0 {
		t.Fatalf("expected nil when context missing model, got %v", res)
	}

	// 4. 上下文具备精准匹配的值
	ctxExact := &mockContext{labels: map[string]string{
		"consumer": "c1",
		"model":    "claude-opus-5",
	}}
	res := matchTree.Match(ctxExact)
	if len(res) != 1 || len(res[0].Items) != 1 || res[0].Items[0] != "strategy_c1" {
		t.Fatalf("expected exact match for strategy_c1, got %+v", res)
	}

	// 5. 上下文给出一个未匹配的 model，由于无通配兜底，返回空
	ctxMismatch := &mockContext{labels: map[string]string{
		"consumer": "c1",
		"model":    "gpt-4o",
	}}
	if res := matchTree.Match(ctxMismatch); len(res) != 0 {
		t.Fatalf("expected no match, got %+v", res)
	}
}
