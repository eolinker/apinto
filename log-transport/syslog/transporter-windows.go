//+build windows
package syslog

import (
	"fmt"

	log_transport "github.com/eolinker/goku/log-transport"

	eosc_log "github.com/eolinker/eosc/log"
)

//CreateTransporter 创建windows下的syslog，windows下不支持syslog，直接返回错误
func CreateTransporter(conf *Config, formatter eosc_log.Formatter) (log_transport.TransporterReset, error) {
	return nil, fmt.Errorf("can not create syslog transporterReset in windows system")
}
