package pricing_policy

import (
	"math"
	"testing"
)

// mockRedisPricingData 生成一个符合 Redis 结构的动态资源定价 mock 对象
func mockRedisPricingData() *RedisPricingData {
	return &RedisPricingData{
		BasicInfo: &BasicInfo{
			Version:         "v1.0.0",
			Rely:            "none",
			ResourceGroupID: "group_ai_standard",
			TenantID:        "tenant_eolinker",
		},
		Strategy: map[string]*PricePlan{
			"base": {
				Cost: map[string]float64{
					"input_token":  3,       // 3 USD / 1M token
					"output_token": 3,       // 3 USD / 1M token
					"per_call":     0.0001,  // 0.0001 USD / call
					"per_second":   0.00001, // 0.00001 USD / second
				},
				Sale: map[string]float64{
					"input_token":  3.5,
					"output_token": 3.5,
					"per_call":     0.003,
					"per_second":   0.003,
				},
				Official: map[string]float64{
					"input_token":  3.5,
					"output_token": 3.5,
					"per_call":     0.003,
					"per_second":   0.003,
				},
			},
			"rule_discount_01": {
				Cost: map[string]float64{
					"input_token":  2,
					"output_token": 2,
					"per_call":     0.00005,
					"per_second":   0.000005,
				},
				Sale: map[string]float64{
					"input_token":  2.5,
					"output_token": 2.5,
					"per_call":     0.002,
					"per_second":   0.002,
				},
				Official: map[string]float64{
					"input_token":  3.5,
					"output_token": 3.5,
					"per_call":     0.003,
					"per_second":   0.003,
				},
			},
			"rule_dalle3_large": {
				Cost: map[string]float64{
					"per_call": 0.08,
				},
				Sale: map[string]float64{
					"per_call": 0.10,
				},
				Official: map[string]float64{
					"per_call": 0.12,
				},
			},
		},
	}
}

// TestCalculator_CalculateToken 测试按 Token 计费的 OpenAI GPT-4 计算器逻辑，绑定 Redis 动态价格并应用兜底 Fallback 逻辑
func TestCalculator_CalculateToken(t *testing.T) {
	config := &Config{
		Currency: "USD",
		AdvancedRules: []*Rule{
			{
				ID:   "rule_discount_01",
				Name: "GPT-4 Token 优惠价规则",
				Conditions: &Condition{
					AllOf: []*AllOf{
						{
							BasicRule: &BasicRule{
								Key:   "input_token",
								Op:    "<=",
								Value: "100000",
								Type:  "integer",
							},
						},
					},
				},
				CostExpression: "(input_token * cost_input_token + output_token * cost_output_token) / 1000000",
				SaleExpression: "(input_token * sale_input_token + output_token * sale_output_token) / 1000000",
			},
			{
				ID:   "rule_standard", // 虽有规则但没有在 Redis Strategy 里定义其专属价格。会 Fallback 退回到 "base" 的价格进行公式运行
				Name: "GPT-4 Token 标准价规则",
				Conditions: &Condition{
					AllOf: []*AllOf{}, // 无条件，代表默认兜底规则
				},
				CostExpression: "(input_token * cost_input_token + output_token * cost_output_token) / 1000000",
				SaleExpression: "(input_token * sale_input_token + output_token * sale_output_token) / 1000000",
			},
		},
	}

	calc, err := NewCalculator(config)
	if err != nil {
		t.Fatalf("failed to create calculator: %v", err)
	}

	redisData := mockRedisPricingData()

	// 1. 验证命中优惠价规则的用量
	vars1 := map[string]interface{}{
		"input_token":  80000,
		"output_token": 20000,
	}
	result1, err := calc.CalculateFromVariables(vars1, redisData)
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	// 预期销售价 (应用 rule_discount_01 售价 2.5): (80000 * 2.5 + 20000 * 2.5) / 1000000 = 250000 / 1000000 = 0.25 USD
	expectedSale1 := 0.25
	if result1.Sale != expectedSale1 {
		t.Errorf("expected sale cost %v, got %v", expectedSale1, result1.Sale)
	}

	// 2. 验证超出优惠阈值，命中兜底规则的用量（会因为规则 rule_standard 没配价格，Fallback 到 base 定价 input_token=3.5 运行）
	vars2 := map[string]interface{}{
		"input_token":  120000,
		"output_token": 50000,
	}
	result2, err := calc.CalculateFromVariables(vars2, redisData)
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	// 预期售价: (120000 * 3.5 + 50000 * 3.5) / 1000000 = (420000 + 175000) / 1000000 = 0.595 USD
	expectedSale2 := 0.595
	if result2.Sale != expectedSale2 {
		t.Errorf("expected sale cost %v, got %v", expectedSale2, result2.Sale)
	}
}

// TestCalculator_CalculateSeconds 测试按秒（语音模型）计费的 Whisper 计算器
func TestCalculator_CalculateSeconds(t *testing.T) {
	config := &Config{
		Currency: "USD",
		AdvancedRules: []*Rule{
			{
				ID:             "rule_discount_01", // 命中 rule_discount_01，使用该专属价格 per_second = 0.002
				Name:           "语音按秒计费规则",
				CostExpression: "duration * cost_per_second",
				SaleExpression: "duration * sale_per_second",
			},
		},
	}

	calc, err := NewCalculator(config)
	if err != nil {
		t.Fatalf("failed to create calculator: %v", err)
	}

	redisData := mockRedisPricingData()
	vars := map[string]interface{}{
		"duration": 120, // 120 秒
	}
	result, err := calc.CalculateFromVariables(vars, redisData)
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}

	// 预期售价: 120 * 0.002 = 0.24 USD
	expectedSale := 0.24
	if result.Sale != expectedSale {
		t.Errorf("expected sale cost %v, got %v", expectedSale, result.Sale)
	}
}

// TestCalculator_CalculateCalls 测试按次（生图模型）计费的 DALL-E 计算器
func TestCalculator_CalculateCalls(t *testing.T) {
	config := &Config{
		Currency: "USD",
		AdvancedRules: []*Rule{
			{
				ID:   "rule_dalle3_large",
				Name: "DALL-E-3 高清分辨率规则",
				Conditions: &Condition{
					AllOf: []*AllOf{
						{
							BasicRule: &BasicRule{
								Key:   "size",
								Op:    "==",
								Value: "1024x1024",
								Type:  "string",
							},
						},
					},
				},
				CostExpression: "call_count * cost_per_call",
				SaleExpression: "call_count * sale_per_call",
			},
			{
				ID:             "rule_standard_calls", // 回退使用 base 售价 per_call = 0.003
				Name:           "DALL-E-3 标准分辨率规则",
				CostExpression: "call_count * cost_per_call",
				SaleExpression: "call_count * sale_per_call",
			},
		},
	}

	calc, err := NewCalculator(config)
	if err != nil {
		t.Fatalf("failed to create calculator: %v", err)
	}

	redisData := mockRedisPricingData()

	// 1. 命中高清 (rule_dalle3_large 售价 0.10)
	vars1 := map[string]interface{}{
		"call_count": 1,
		"size":       "1024x1024",
	}
	result1, err := calc.CalculateFromVariables(vars1, redisData)
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}
	if result1.Sale != 0.10 {
		t.Errorf("expected 0.10, got %v", result1.Sale)
	}

	// 2. 命中标准 (Fallback to base, 售价 0.003)
	vars2 := map[string]interface{}{
		"call_count": 3,
		"size":       "512x512",
	}
	result2, err := calc.CalculateFromVariables(vars2, redisData)
	if err != nil {
		t.Fatalf("calculation failed: %v", err)
	}
	// 预期售价: 3 * 0.003 = 0.009
	if math.Abs(result2.Sale-0.009) > 1e-9 {
		t.Errorf("expected 0.009, got %v", result2.Sale)
	}
}

// TestCalculator_MatchConditions 测试 AllOf、AnyOf 与 in 操作符复合匹配
func TestCalculator_MatchConditions(t *testing.T) {
	// 测试 in 操作符与整数
	ruleInInt := &BasicRule{
		Key:   "status_code",
		Op:    "in",
		Value: "200, 201, 204",
		Type:  "integer",
	}
	paramsPassInt := map[string]interface{}{"status_code": 201}
	paramsFailInt := map[string]interface{}{"status_code": 404}

	if !matchBasicRule(ruleInInt, paramsPassInt) {
		t.Error("expected true for status_code 201 in [200, 201, 204]")
	}
	if matchBasicRule(ruleInInt, paramsFailInt) {
		t.Error("expected false for status_code 404 in [200, 201, 204]")
	}

	// 测试 in 操作符与字符串
	ruleInStr := &BasicRule{
		Key:   "size",
		Op:    "in",
		Value: "1024x1024, 512x512",
		Type:  "string",
	}
	paramsPassStr := map[string]interface{}{"size": "512x512"}
	paramsFailStr := map[string]interface{}{"size": "256x256"}

	if !matchBasicRule(ruleInStr, paramsPassStr) {
		t.Error("expected true for size 512x512 in [1024x1024, 512x512]")
	}
	if matchBasicRule(ruleInStr, paramsFailStr) {
		t.Error("expected false for size 256x256 in [1024x1024, 512x512]")
	}

	// 测试 AnyOf 包含嵌套 AllOf 的递归评估，来表达: "status == 200" OR ("status == 302" AND "method == GET")
	cond := &Condition{
		AnyOf: []*AnyOf{
			{
				BasicRule: &BasicRule{
					Key:   "status",
					Op:    "==",
					Value: "200",
					Type:  "integer",
				},
			},
			{
				AllOf: []*AllOf{
					{
						BasicRule: &BasicRule{
							Key:   "status",
							Op:    "==",
							Value: "302",
							Type:  "integer",
						},
						AnyOf: []*AnyOf{
							{
								BasicRule: &BasicRule{
									Key:   "method",
									Op:    "==",
									Value: "GET",
									Type:  "string",
								},
							},
						},
					},
				},
			},
		},
	}

	// status 200 符合 AnyOf 第一个分支
	if !matchCondition(cond, map[string]interface{}{"status": 200}) {
		t.Error("expected true for status 200")
	}

	// status 302 且 method GET 符合 AnyOf 第二个分支嵌套 (status == 302 并且其 any_of 下 method == GET 也符合)
	if !matchCondition(cond, map[string]interface{}{"status": 302, "method": "GET"}) {
		t.Error("expected true for status 302 with GET method")
	}

	// status 302 但 method POST 不符合嵌套子 AnyOf 中的条件，应该被拒绝
	if matchCondition(cond, map[string]interface{}{"status": 302, "method": "POST"}) {
		t.Error("expected false for status 302 with POST method")
	}
}
