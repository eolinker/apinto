package price_calcular

type Rule struct {
	ID                 string            `json:"id" label:"规则ID"`
	Name               string            `json:"name" label:"规则名称"`
	Conditions         *Condition        `json:"conditions" label:"条件"`
	Price              map[string]string `json:"price" label:"价格"`
	CostExpression     string            `json:"cost_expression" label:"进货价计算表达式"`
	SaleExpression     string            `json:"sale_expression" label:"销售价计算表达式"`
	OfficialExpression string            `json:"official_expression" label:"官方价计算表达式"`
}

type Condition struct {
	AllOf []*BasicRule `json:"all_of" label:"满足全部条件，只能二选一"`
	OneOf []*BasicRule `json:"one_of" label:"满足任意条件，只能二选一"`
}

type BasicRule struct {
	Key   string `json:"key" label:"变量key"`
	Op    string `json:"op" label:"运算符" enum:">,>=,==,<,<=,!=,in"`
	Value string `json:"value" label:"值"`
	Type  string `json:"type" label:"类型" enum:"string,integer,float,boolean"`
}
