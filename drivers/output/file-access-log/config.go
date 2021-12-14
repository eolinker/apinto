package file_access_log

import (
	"errors"

	"github.com/eolinker/eosc/formatter"
)

var (
	errorConfigType = errors.New("error config type")
)

type Config struct {
	File      string           `json:"file" yaml:"file"`
	Dir       string           `json:"dir" yaml:"dir"`
	Period    string           `json:"period" yaml:"period"`
	Expire    int              `json:"expire" yaml:"expire"`
	Type      string           `json:"type" yaml:"type"`
	Formatter formatter.Config `json:"formatter" yaml:"formatter"`
}
