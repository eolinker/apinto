package price_calcular

import (
	"testing"
)

func TestPricingData_MaxSaleStrategyID(t *testing.T) {
	data := &PricingData{
		Strategy: map[string]*PricePlan{
			"strategy_low": {
				Sale: map[string]float64{
					"input_token":  1.0,
					"output_token": 2.0,
				},
			},
			"strategy_high": {
				Sale: map[string]float64{
					"input_token":  5.0,
					"output_token": 10.0,
				},
			},
			"strategy_mid": {
				Sale: map[string]float64{
					"input_token":  3.0,
					"output_token": 4.0,
				},
			},
		},
	}

	maxID := data.MaxSaleStrategyID()
	if maxID != "strategy_high" {
		t.Errorf("MaxSaleStrategyID got %q, want %q", maxID, "strategy_high")
	}

	var nilData *PricingData
	if nilData.MaxSaleStrategyID() != "" {
		t.Errorf("MaxSaleStrategyID on nil got %q, want empty string", nilData.MaxSaleStrategyID())
	}

	emptyData := &PricingData{}
	if emptyData.MaxSaleStrategyID() != "" {
		t.Errorf("MaxSaleStrategyID on empty got %q, want empty string", emptyData.MaxSaleStrategyID())
	}
}
