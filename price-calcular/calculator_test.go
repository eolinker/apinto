package price_calcular

import (
	"testing"
)

func TestCalculator_CalculateByRule(t *testing.T) {
	variables := Variables{
		"input_token": {
			Source: "label",
			Type:   "integer",
		},
		"output_token": {
			Source: "label",
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
		},
	}
	
	calc, err := NewCalculator("aa", "USD", variables, rules)
	if err != nil {
		t.Fatalf("NewCalculator failed: %v", err)
	}
	
	pricingData := &PricingData{
		Strategy: map[string]*PricePlan{
			"rule_001": {
				Cost: map[string]float64{
					"input_token":  1.0,
					"output_token": 2.0,
				},
				Sale: map[string]float64{
					"input_token":  2.0,
					"output_token": 4.0,
				},
				Official: map[string]float64{
					"input_token":  3.0,
					"output_token": 6.0,
				},
			},
		},
	}
	
	vars := map[string]interface{}{
		"input_token":  int64(1000000), // 1M tokens
		"output_token": int64(500000),  // 0.5M tokens
	}
	
	res, err := calc.CalculateByRule(nil, "rule_001", vars, pricingData)
	if err != nil {
		t.Fatalf("CalculateByRule error: %v", err)
	}
	
	// Cost = 1.0 * 1M/1M + 2.0 * 0.5M/1M = 1.0 + 1.0 = 2.0
	expectedCost := 2.0
	// Sale = 2.0 * 1M/1M + 4.0 * 0.5M/1M = 2.0 + 2.0 = 4.0
	expectedSale := 4.0
	// Official = 3.0 * 1M/1M + 6.0 * 0.5M/1M = 3.0 + 3.0 = 6.0
	expectedOfficial := 6.0
	
	if res.Cost != expectedCost {
		t.Errorf("Cost got %f, want %f", res.Cost, expectedCost)
	}
	if res.Sale != expectedSale {
		t.Errorf("Sale got %f, want %f", res.Sale, expectedSale)
	}
	if res.Official != expectedOfficial {
		t.Errorf("Official got %f, want %f", res.Official, expectedOfficial)
	}
}
