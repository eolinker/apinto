package file_access_log

import "github.com/eolinker/eosc/formatter"

type Config struct {
	Dir       string           `json:"dir" yaml:"dir"`
	Period    string           `json:"period" yaml:"period"`
	Expire    int              `json:"expire" yaml:"expire"`
	Type      string           `json:"tyoe" yaml:"tyoe"`
	Formatter formatter.Config `json:"formatter" yaml:"formatter"`
}
