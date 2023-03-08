package protocbuf

import (
	"encoding/json"

	"github.com/eolinker/eosc"
)

// Config protobuf驱动配置
type Config struct {
	ProtoFiles eosc.EoFiles `json:"proto_files" label:"proto文件列表"`
}

func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}
