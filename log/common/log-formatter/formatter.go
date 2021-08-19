package log_config_module

import (
	"github.com/eolinker/eosc/log"
	formatter_json "github.com/eolinker/goku/log/common/formatter/formatter-json"
	"strings"
)

var (
	allFormatterName = map[string]bool{
		"json": true,
		"line": true,
	}
)

func CreateFormatter(formatterName string) log.Formatter {
	formatterName = strings.ToLower(formatterName)
	if !allFormatterName[formatterName] {

	}
	if formatterName == "" {
		formatterName = "line"
	}

	switch strings.ToLower(formatterName) {
	case "json":
		return &formatter_json.JSONFormatter{}
	case "line":
		fallthrough
	default:

		return &log.LineFormatter{
			TimestampFormat:  "[2006-01-02 15:04:05]",
			CallerPrettyfier: nil,
		}
	}
}
