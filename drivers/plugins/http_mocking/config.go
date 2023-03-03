package http_mocking

type Config struct {
	ResponseStatus  int               `json:"response_status" default:"200" label:"返回响应的 HTTP 状态码（仅http路由有效）"`
	ContentType     string            `json:"content_type" label:"返回响应的 Header Content-Type" enum:"application/json"`
	ResponseExample string            `json:"response_example"  format:"text" label:"返回响应的Body，与jsonschema字段二选一"`
	ResponseSchema  string            `json:"response_schema" format:"text" label:"指定响应的jsonschema对象"`
	ResponseHeader  map[string]string `json:"response_header" label:"响应头"`
}

const (
	contentTypeJson = "application/json"
)
