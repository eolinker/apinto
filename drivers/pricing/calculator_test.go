package pricing_driver

import (
	"testing"

	"github.com/eolinker/apinto/pricing"
)

func TestCalculate_Token(t *testing.T) {
	pc := NewPriceCalculator("openai", pricing.ResourcePricing{
		Resource:    "gpt-4",
		Mode:        pricing.PricingModeToken,
		Input:       3.0,
		Output:      6.0,
		CachedInput: 1.5,
	}, pricing.PricingPhaseSync, nil, nil)

	got, err := pc.Calculate(&pricing.PricingUsage{
		InputCount:  1_000_000,
		OutputCount: 500_000,
		CachedCount: 200_000,
	})
	if err != nil {
		t.Fatalf("calculate err: %v", err)
	}
	// input=3.0, output=6/1M*500k=3.0, cached=1.5/1M*200k=0.3
	wantTotal := 3.0 + 3.0 + 0.3
	if got.TotalCost != wantTotal {
		t.Fatalf("want total %.4f, got %.4f", wantTotal, got.TotalCost)
	}
	if got.ChargeType != pricing.ChargeTypeActual {
		t.Fatalf("want actual, got %s", got.ChargeType)
	}
}

func TestCalculate_PerCall(t *testing.T) {
	pc := NewPriceCalculator("svc", pricing.ResourcePricing{
		Resource: "/api/order",
		Mode:     pricing.PricingModePerCall,
		PerCall:  0.05,
	}, pricing.PricingPhaseSync, nil, nil)

	got, err := pc.Calculate(&pricing.PricingUsage{CallCount: 3})
	if err != nil {
		t.Fatalf("calculate err: %v", err)
	}
	if got.TotalCost != 0.15 {
		t.Fatalf("want 0.15, got %.4f", got.TotalCost)
	}

	// CallCount==0 应当被视为 1 次
	got2, _ := pc.Calculate(&pricing.PricingUsage{})
	if got2.TotalCost != 0.05 {
		t.Fatalf("want 0.05 default call, got %.4f", got2.TotalCost)
	}
}

func TestCalculate_PerSecond(t *testing.T) {
	pc := NewPriceCalculator("svc", pricing.ResourcePricing{
		Resource:  "video",
		Mode:      pricing.PricingModePerSecond,
		PerSecond: 0.02,
	}, pricing.PricingPhaseSync, nil, nil)

	got, err := pc.Calculate(&pricing.PricingUsage{Duration: 12.5})
	if err != nil {
		t.Fatalf("calculate err: %v", err)
	}
	want := 0.02 * 12.5
	if got.TotalCost != want {
		t.Fatalf("want %.4f, got %.4f", want, got.TotalCost)
	}
}

func TestCalculate_Submit_PreCharge(t *testing.T) {
	pc := NewPriceCalculator("svc", pricing.ResourcePricing{
		Resource: "task",
		Mode:     pricing.PricingModePerCall,
		PerCall:  0.5,
	}, pricing.PricingPhaseSubmit, nil, nil)

	got, _ := pc.Calculate(&pricing.PricingUsage{CallCount: 1})
	if got.ChargeType != pricing.ChargeTypePre {
		t.Fatalf("submit phase want pre charge, got %s", got.ChargeType)
	}
}

func TestCalculate_AdvancedMatch(t *testing.T) {
	pc := NewPriceCalculator("svc", pricing.ResourcePricing{
		Resource: "/api/order",
		Mode:     pricing.PricingModeAdvanced,
		PerCall:  0.10, // 默认价
		AdvancedRules: []pricing.AdvancedRule{
			{
				Condition: pricing.Condition{Type: pricing.ConditionStatusCode, Operator: ">=", Threshold: 500},
				Pricing:   pricing.TierPricing{PerCall: 0}, // 5xx 免费
			},
			{
				Condition: pricing.Condition{Type: pricing.ConditionStatusCode, Operator: ">=", Threshold: 400},
				Pricing:   pricing.TierPricing{PerCall: 0.02}, // 4xx 减价
			},
		},
	}, pricing.PricingPhaseSync, nil, nil)

	// 5xx 命中第一条
	r1, _ := pc.Calculate(&pricing.PricingUsage{StatusCode: 502, CallCount: 1})
	if r1.TotalCost != 0 {
		t.Fatalf("5xx want 0, got %.4f", r1.TotalCost)
	}
	// 4xx 命中第二条
	r2, _ := pc.Calculate(&pricing.PricingUsage{StatusCode: 404, CallCount: 1})
	if r2.TotalCost != 0.02 {
		t.Fatalf("4xx want 0.02, got %.4f", r2.TotalCost)
	}
	// 2xx 全部不匹配，回退基础 PerCall
	r3, _ := pc.Calculate(&pricing.PricingUsage{StatusCode: 200, CallCount: 1})
	if r3.TotalCost != 0.10 {
		t.Fatalf("2xx fallback want 0.10, got %.4f", r3.TotalCost)
	}
}

func TestMatchCondition_MethodAndPath(t *testing.T) {
	usage := &pricing.PricingUsage{Method: "POST", Path: "/api/v1/users/123"}
	if !matchCondition(&pricing.Condition{Type: pricing.ConditionMethod, Value: "post"}, usage) {
		t.Fatalf("method case-insensitive failed")
	}
	if !matchCondition(&pricing.Condition{Type: pricing.ConditionPath, Value: "/api/v1/users/*"}, usage) {
		t.Fatalf("path wildcard failed")
	}
	if matchCondition(&pricing.Condition{Type: pricing.ConditionPath, Value: "/api/v2/*"}, usage) {
		t.Fatalf("unmatched path should be false")
	}
	if !matchCondition(&pricing.Condition{Type: pricing.ConditionPath, Value: "/api/v1/users/123"}, usage) {
		t.Fatalf("path exact failed")
	}
}

func TestApplyDiscount_Official(t *testing.T) {
	base := pricing.ResourcePricing{
		Resource: "x",
		Mode:     pricing.PricingModeToken,
		Input:    10, Output: 20, CachedInput: 5, PerCall: 0.5, PerSecond: 0.1,
		AdvancedRules: []pricing.AdvancedRule{
			{Pricing: pricing.TierPricing{Input: 4, Output: 8}},
		},
	}
	got := applyDiscount(base, &DynamicPricingStrategy{
		ReferencePrice: pricing.PriceLevelOfficial,
		Discount:       0.5,
	})
	if got.Input != 5 || got.Output != 10 {
		t.Fatalf("base discount failed: %+v", got)
	}
	if got.AdvancedRules[0].Pricing.Input != 2 || got.AdvancedRules[0].Pricing.Output != 4 {
		t.Fatalf("tier discount failed: %+v", got.AdvancedRules[0].Pricing)
	}
	// 原配置不应被修改
	if base.Input != 10 {
		t.Fatalf("base mutated: %.2f", base.Input)
	}
}

func TestCompareThreshold(t *testing.T) {
	cases := []struct {
		v, t float64
		op   string
		want bool
	}{
		{10, 5, ">", true},
		{5, 5, ">=", true},
		{4, 5, "<", true},
		{5, 5, "<=", true},
		{5, 5, "==", true},
		{5, 5, "=", true},
		{5, 5, "!=", false},
	}
	for _, c := range cases {
		if got := compareThreshold(c.v, c.op, c.t); got != c.want {
			t.Fatalf("compare %g %s %g want %v got %v", c.v, c.op, c.t, c.want, got)
		}
	}
}
