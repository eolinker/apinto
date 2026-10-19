package price_calcular

import (
	"context"
	"testing"

	context_label "github.com/eolinker/apinto/common/context-label"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type mockResp struct {
	http_context.IResponse
	body []byte
}

func (m *mockResp) GetBody() []byte {
	return m.body
}

type mockContext struct {
	http_context.IHttpContext
	labels map[string]string
	values map[string]interface{}
	resp   *mockResp
}

func newMockContext(body []byte) *mockContext {
	return &mockContext{
		labels: make(map[string]string),
		values: make(map[string]interface{}),
		resp:   &mockResp{body: body},
	}
}

func (m *mockContext) Context() context.Context {
	return context.Background()
}

func (m *mockContext) Assert(i interface{}) error {
	if v, ok := i.(*http_context.IHttpContext); ok {
		*v = m
		return nil
	}
	return nil
}

func (m *mockContext) Value(key interface{}) interface{} {
	if k, ok := key.(string); ok {
		return m.values[k]
	}
	return nil
}

func (m *mockContext) WithValue(key, val interface{}) {
	if k, ok := key.(string); ok {
		m.values[k] = val
	}
}

func (m *mockContext) GetLabel(key string) string {
	return m.labels[key]
}

func (m *mockContext) SetLabel(key, val string) {
	m.labels[key] = val
}

func (m *mockContext) Response() http_context.IResponse {
	return m.resp
}

func TestCalculator_Calculate(t *testing.T) {
	variables := Variables{
		"input_token": {
			Source: "response_body",
			Path:   "usage.prompt_tokens",
			Type:   "integer",
		},
		"output_token": {
			Source: "response_body",
			Path:   "usage.completion_tokens",
			Type:   "integer",
		},
	}

	rules := []*Rule{
		{
			ID:                 "rule_001",
			Name:               "通用 token 计费规则",
			CostExpression:     "cost_input_token * input_token / 1000000 + cost_output_token * output_token / 1000000",
			SaleExpression:     "sale_input_token * input_token / 1000000 + sale_output_token * output_token / 1000000",
			OfficialExpression: "official_input_token * input_token / 1000000 + official_output_token * output_token / 1000000",
			MarkedExpression:   "marked_input_token * input_token / 1000000 + marked_output_token * output_token / 1000000",
		},
	}

	calc, err := NewCalculator("aa", "USD", variables, rules)
	if err != nil {
		t.Fatalf("NewCalculator failed: %v", err)
	}

	ruleList, err := calc.Rules(nil)
	if err != nil || len(ruleList) == 0 {
		t.Fatalf("calc.Rules failed: %v", err)
	}

	matchedRule := ruleList[0]

	params := map[string]interface{}{
		"input_token":           int64(1000000), // 1M tokens
		"output_token":          int64(500000),  // 0.5M tokens
		"cost_input_token":      1.0,
		"cost_output_token":     2.0,
		"sale_input_token":      2.0,
		"sale_output_token":     4.0,
		"official_input_token":  3.0,
		"official_output_token": 6.0,
		"marked_input_token":    5.0,
		"marked_output_token":   10.0,
	}

	res, err := calc.(*Calculator).CalculateFromVariables(params, matchedRule)
	if err != nil {
		t.Fatalf("CalculateFromVariables error: %v", err)
	}

	// Cost = 1.0 * 1M/1M + 2.0 * 0.5M/1M = 1.0 + 1.0 = 2.0
	expectedCost := 2.0
	// Sale = 2.0 * 1M/1M + 4.0 * 0.5M/1M = 2.0 + 2.0 = 4.0
	expectedSale := 4.0
	// Official = 3.0 * 1M/1M + 6.0 * 0.5M/1M = 3.0 + 3.0 = 6.0
	expectedOfficial := 6.0
	// Marked = 5.0 * 1M/1M + 10.0 * 0.5M/1M = 5.0 + 5.0 = 10.0
	expectedMarked := 10.0

	if res.Cost != expectedCost {
		t.Errorf("Cost got %f, want %f", res.Cost, expectedCost)
	}
	if res.Sale != expectedSale {
		t.Errorf("Sale got %f, want %f", res.Sale, expectedSale)
	}
	if res.Official != expectedOfficial {
		t.Errorf("Official got %f, want %f", res.Official, expectedOfficial)
	}
	if res.Marked != expectedMarked {
		t.Errorf("Marked got %f, want %f", res.Marked, expectedMarked)
	}
}

func TestCalculator_GeminiUsageBilling(t *testing.T) {
	variables := Variables{
		"input_token": {
			Source: "response_body",
			Path:   "$.usageMetadata.promptTokenCount",
			Type:   "integer",
		},
		"output_token": {
			Source: "response_body",
			Path:   "$.usageMetadata.candidatesTokenCount",
			Type:   "integer",
		},
		"cache_token": {
			Source: "response_body",
			Path:   "$.usageMetadata.cachedContentTokenCount",
			Type:   "integer",
		},
		"thoughts_token": {
			Source: "response_body",
			Path:   "$.usageMetadata.thoughtsTokenCount",
			Type:   "integer",
		},
		"output_text_token": {
			Source: "response_body",
			Path:   "$.usageMetadata.candidatesTokensDetails[?(@.modality=='TEXT')].tokenCount",
			Type:   "integer",
		},
		"output_image_token": {
			Source: "response_body",
			Path:   "$.usageMetadata.candidatesTokensDetails[?(@.modality=='IMAGE')].tokenCount",
			Type:   "integer",
		},
		"total_token": {
			Source: "response_body",
			Path:   "$.usageMetadata.totalTokenCount",
			Type:   "integer",
		},
	}

	rules := []*Rule{
		{
			ID:             "base",
			Name:           "基本计费规则",
			CostExpression: "(input_token * #cost.input_token + cache_token * #cost.cache_token + (output_text_token + thoughts_token) * #cost.output_text_token + output_image_token * #cost.output_image_token) / 1000000",
			SaleExpression: "(input_token * #sale.input_token + cache_token * #sale.cache_token + (output_text_token + thoughts_token) * #sale.output_text_token + output_image_token * #sale.output_image_token) / 1000000",
		},
	}

	calc, err := NewCalculator("gemini_calc", "USD", variables, rules)
	if err != nil {
		t.Fatalf("NewCalculator failed: %v", err)
	}

	rawResp := []byte(`{
  "usageMetadata": {
    "promptTokenCount": 41,
    "candidatesTokenCount": 1120,
    "totalTokenCount": 2040,
    "trafficType": "ON_DEMAND",
    "promptTokensDetails": [
      {
        "modality": "TEXT",
        "tokenCount": 41
      }
    ],
    "candidatesTokensDetails": [
      {
        "modality": "IMAGE",
        "tokenCount": 1120
      }
    ],
    "thoughtsTokenCount": 879
  }
}`)

	_ = calc
	ctx := newMockContext(rawResp)
	extractor := calc.VariablesExtractor()
	vars := extractor.ExtractAll(ctx)

	expectedVars := map[string]int64{
		"input_token":        41,
		"output_token":       1120,
		"cache_token":        0,
		"thoughts_token":     879,
		"output_text_token":  0,
		"output_image_token": 1120,
		"total_token":        2040,
	}

	for k, expected := range expectedVars {
		actual, ok := vars[k]
		if !ok {
			t.Errorf("missing variable %s", k)
			continue
		}
		if actual != expected {
			t.Errorf("variable %s got %v (%T), want %v", k, actual, actual, expected)
		}
	}

	pricingData := &PricingData{
		Strategy: map[string]*PricePlan{
			"base": {
				Cost: map[string]float64{
					"input_token":        1.0,
					"cache_token":        0.5,
					"output_text_token":  2.0,
					"output_image_token": 10.0,
				},
				Sale: map[string]float64{
					"input_token":        2.0,
					"cache_token":        1.0,
					"output_text_token":  4.0,
					"output_image_token": 20.0,
				},
			},
		},
	}

	res, err := calc.Calculate(ctx, true, pricingData)
	if err != nil {
		t.Fatalf("calc.Calculate failed: %v", err)
	}

	// Cost: (41 * 1.0 + 0 * 0.5 + (0 + 879) * 2.0 + 1120 * 10.0) / 1000000 = (41 + 1758 + 11200) / 1000000 = 0.012999
	expectedCost := 0.012999
	// Sale: (41 * 2.0 + 0 * 1.0 + (0 + 879) * 4.0 + 1120 * 20.0) / 1000000 = (82 + 3516 + 22400) / 1000000 = 0.025998
	expectedSale := 0.025998

	if res.Cost != expectedCost {
		t.Errorf("Cost got %f, want %f", res.Cost, expectedCost)
	}
	if res.Sale != expectedSale {
		t.Errorf("Sale got %f, want %f", res.Sale, expectedSale)
	}

	if gotUnit := context_label.GetPriceUnit(ctx); gotUnit != BillingModeToken {
		t.Errorf("price_unit got %q, want %q", gotUnit, BillingModeToken)
	}
	wantCostNumExpr := "(41 * 1 + 0 * 0.5 + (0 + 879) * 2 + 1120 * 10) / 1000000"
	if gotCostNumExpr := context_label.GetNumExprCost(ctx); gotCostNumExpr != wantCostNumExpr {
		t.Errorf("cost_num_expr got %q, want %q", gotCostNumExpr, wantCostNumExpr)
	}
	wantSaleNumExpr := "(41 * 2 + 0 * 1 + (0 + 879) * 4 + 1120 * 20) / 1000000"
	if gotSaleNumExpr := context_label.GetNumExprSale(ctx); gotSaleNumExpr != wantSaleNumExpr {
		t.Errorf("sale_num_expr got %q, want %q", gotSaleNumExpr, wantSaleNumExpr)
	}
}
