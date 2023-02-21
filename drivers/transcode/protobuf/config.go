package protocbuf

import (
	"encoding/json"
)

// Config service_http驱动配置
type Config struct {
}

func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}
