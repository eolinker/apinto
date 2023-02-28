package http_mocking

type Config struct {
	ResponseStatus  int    `json:"response_status" default:"200" label:"返回响应的 HTTP 状态码（仅http路由有效）"`
	ContentType     string `json:"content_type" label:"返回响应的 Header Content-Type" enum:"application/json"`
	ResponseExample string `json:"response_example"  switch:"content_type==='application/json'" label:"返回响应的Body，与jsonschema字段二选一"`
	ResponseSchema  string `json:"response_schema"  switch:"content_type==='application/json'" label:"指定响应的jsonschema对象"`
}

const (
	contentTypeJson = "application/json"
)
