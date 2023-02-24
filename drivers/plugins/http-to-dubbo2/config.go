package http_to_dubbo2

type Config struct {
	Service string   `json:"service" label:"服务名称" required:"true"`
	Method  string   `json:"method" label:"方法名称" required:"true"`
	Params  []*Param `json:"params" label:"参数" required:"true"`
}

type Param struct {
	ClassName string `json:"class_name" label:"class_name" required:"true"` //对应Java中类的class_name
	FieldName string `json:"field_name" label:"根字段名"`                       //读取body中json的根字段名
}
