package params_check

type Config struct {
	Params []*Param `json:"params" label:"参数列表"`
}

type Param struct {
	Name      string `json:"name" label:"参数名"`
	Position  string `json:"position" label:"参数位置" enum:"query,header,body"`
	MatchText string `json:"match_text" label:"匹配文本"`
}
