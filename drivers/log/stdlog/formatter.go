package stdlog

import (
	"fmt"
	"strings"

	"github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/log-transport/common/formatter/json"
)

var (
	allFormatterName = map[string]bool{
		"json": true,
		"line": true,
	}
)

//CreateFormatter 创建输出格式处理器
func CreateFormatter(formatterName string) (log.Formatter, error) {
	if formatterName == "" {
		formatterName = "line"
	}

	formatterName = strings.ToLower(formatterName)
	if !allFormatterName[formatterName] {
		return nil, fmt.Errorf("formatterName:%s is not supported", formatterName)
	}

	switch strings.ToLower(formatterName) {
	case "json":
		return &json.JSONFormatter{}, nil
	case "line":
		fallthrough
	default:
		return &log.LineFormatter{
			TimestampFormat:  "[2006-01-02 15:04:05]",
			CallerPrettyfier: nil,
		}, nil
	}
}
