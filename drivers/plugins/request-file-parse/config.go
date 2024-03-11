package request_file_parse

type Config struct {
	FileKey       string   `json:"file_key" label:"文件Key"`
	FileSuffix    []string `json:"file_suffix" label:"文件有效后缀列表"`
	LargeWarn     int64    `json:"large_warn" label:"文件大小警告阈值"`
	LargeWarnText string   `json:"large_warn_text" label:"文件大小警告标签值"`
}
