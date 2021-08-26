package filelog

import (
	"fmt"
	"strings"

	"github.com/eolinker/goku/log-transport/common/formatter/json"

	"github.com/eolinker/eosc/log"
)

var (
	allFormatterName = map[string]bool{
		"json": true,
		"line": true,
	}
)

func CreateFormatter(formatterName string) (log.Formatter, error) {
	if formatterName == "" {
		formatterName = "line"
	}

	formatterName = strings.ToLower(formatterName)
	if formatterName != "" && !allFormatterName[formatterName] {
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
