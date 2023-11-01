package body_record_truncation

type Config struct {
	BodySize int64 `json:"body_size" label:"截断大小"`
}
