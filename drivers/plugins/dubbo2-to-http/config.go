package dubbo2_to_http

type Config struct {
	Method      string  `json:"method" label:"方法"  enum:"POST,GET,HEAD,PUT,PATCH,DELETE,CONNECT,OPTIONS,TRACE" required:"true"` //get/post
	Path        string  `json:"path" label:"转发路径" required:"true"`
	ContentType string  `json:"content_type" label:"ContentType" enum:"application/json" required:"true"` // application/json 只支持json格式传输
	Params      []Param `json:"params" label:"参数解析" required:"true"`
}
type Param struct {
	ClassName string `json:"class_name" label:"class_name" required:"true"` //对应Java中类的class_name
	FieldName string `json:"field_name" label:"字段名"`                        //用于http传输中body中的key名
}
