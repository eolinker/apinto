package pricing

// CalculatorSkill 价格计算器的技能标识，用于 worker 之间的引用约定。
const CalculatorSkill = "github.com/eolinker/apinto/pricing.IPriceCalculator"

// CheckCalculatorSkill 判断 skill 字符串是否为价格计算器技能。
func CheckCalculatorSkill(skill string) bool {
	return skill == CalculatorSkill
}
