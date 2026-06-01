package billing

import (
	"testing"

	"github.com/eolinker/apinto/pricing"
)

func TestMatchValue_StringNumberBool(t *testing.T) {
	if !matchValue("ok", "ok") {
		t.Fatalf("string equal failed")
	}
	if matchValue("ok", "err") {
		t.Fatalf("string non-equal should be false")
	}
	if !matchValue(0, 0) {
		t.Fatalf("int equal failed")
	}
	if !matchValue(1.5, 1.5) {
		t.Fatalf("float equal failed")
	}
	if !matchValue(true, true) {
		t.Fatalf("bool equal failed")
	}
	if matchValue(true, false) {
		t.Fatalf("bool non-equal should be false")
	}
	// 类型不一致退化为字符串比较
	if !matchValue(123, "123") {
		t.Fatalf("string fallback failed")
	}
}

func TestMatchFieldRuleValue_OperatorVariants(t *testing.T) {
	// 空 / "==" / "=" 走精确匹配
	if !matchFieldRuleValue("ok", pricing.ResultFieldRule{Field: "code", Value: "ok"}) {
		t.Fatalf("empty operator equal failed")
	}
	if !matchFieldRuleValue(200, pricing.ResultFieldRule{Field: "code", Value: 200, Operator: "=="}) {
		t.Fatalf("== operator equal failed")
	}

	// 数值阈值
	if !matchFieldRuleValue(500, pricing.ResultFieldRule{Field: "x", Value: 400, Operator: ">="}) {
		t.Fatalf(">= failed")
	}
	if matchFieldRuleValue(300, pricing.ResultFieldRule{Field: "x", Value: 400, Operator: ">="}) {
		t.Fatalf(">= 300>=400 should be false")
	}
	// 实际值非数值
	if matchFieldRuleValue("foo", pricing.ResultFieldRule{Field: "x", Value: 400, Operator: ">="}) {
		t.Fatalf("non-numeric actual should fail >=")
	}
	// 期望值非数值
	if matchFieldRuleValue(500, pricing.ResultFieldRule{Field: "x", Value: "foo", Operator: ">="}) {
		t.Fatalf("non-numeric expected should fail >=")
	}
}

func TestResultCompareThreshold(t *testing.T) {
	cases := []struct {
		v, t float64
		op   string
		want bool
	}{
		{10, 5, ">", true},
		{5, 5, "<=", true},
		{4, 5, "<", true},
		{6, 5, ">=", true},
		{5, 5, "==", true},
		{5, 5, "=", true},
		{5, 4, "!=", false}, // 不支持的运算符
	}
	for _, c := range cases {
		if got := resultCompareThreshold(c.v, c.op, c.t); got != c.want {
			t.Fatalf("compare %g %s %g want %v got %v", c.v, c.op, c.t, c.want, got)
		}
	}
}

func TestToFloat(t *testing.T) {
	cases := []struct {
		in   interface{}
		ok   bool
		want float64
	}{
		{1.5, true, 1.5},
		{int(7), true, 7},
		{int64(11), true, 11},
		{"foo", false, 0},
		{nil, false, 0},
	}
	for _, c := range cases {
		got, ok := toFloat(c.in)
		if ok != c.ok || got != c.want {
			t.Fatalf("toFloat(%v) want (%v,%v), got (%v,%v)", c.in, c.want, c.ok, got, ok)
		}
	}
}

func TestResultMatcher_NilMatchesAlways(t *testing.T) {
	var m *resultMatcher
	if !m.Match(nil, nil) {
		t.Fatalf("nil matcher should match")
	}
}
