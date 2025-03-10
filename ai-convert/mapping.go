package ai_convert

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ValueRule struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

// MappingRule 定义映射规则结构
type MappingRule map[string]*ValueRule

// TransformData 执行字段映射和类型转换
func TransformData(inputJSON string, mappingRule MappingRule) (map[string]interface{}, error) {
	// 1. 解析输入的JSON字符串到map
	var inputMap map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &inputMap); err != nil {
		return nil, fmt.Errorf("解析输入JSON失败: %v", err)
	}

	// 2. 创建结果map
	resultMap := make(map[string]interface{})

	// 3. 执行字段映射和类型转换
	for oldKey, rule := range mappingRule {
		if value, exists := inputMap[oldKey]; exists {
			// 根据目标类型进行转换
			convertedValue, err := convertType(value, rule.Type)
			if err != nil {
				return nil, fmt.Errorf("类型转换失败 %s -> %s: %v", oldKey, rule.Value, err)
			}
			resultMap[rule.Value] = convertedValue
		} else {
			// 如果源字段不存在，直接复制
			resultMap[rule.Value] = inputMap[oldKey]
		}
	}

	return resultMap, nil
}

// convertType 处理类型转换
func convertType(value interface{}, targetType string) (interface{}, error) {
	switch targetType {
	case "string":
		return fmt.Sprintf("%v", value), nil
	case "number", "float64":
		switch v := value.(type) {
		case string:
			if num, err := strconv.ParseFloat(v, 64); err == nil {
				return num, nil
			}
			return nil, fmt.Errorf("无法将字符串转换为数字: %v", value)
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		default:
			return nil, fmt.Errorf("不支持的数字类型: %T", value)
		}
	case "int":
		switch v := value.(type) {
		case string:
			if num, err := strconv.Atoi(v); err == nil {
				return num, nil
			}
			return nil, fmt.Errorf("无法将字符串转换为整数: %v", value)
		case int:
			return v, nil
		case float64:
			return int(v), nil
		default:
			return nil, fmt.Errorf("不支持的整数类型: %T", value)
		}
	case "boolean":
		switch v := value.(type) {
		case string:
			return strconv.ParseBool(v)
		case bool:
			return v, nil
		default:
			return nil, fmt.Errorf("不支持的布尔类型: %T", value)
		}
	default:
		return value, nil // 默认保持原类型
	}
}
