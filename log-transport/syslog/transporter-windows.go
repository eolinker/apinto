//+build windows
package syslog

import (
	"fmt"
	eosc_log "github.com/eolinker/eosc/log"
)

type TransporterWindows struct {
	*eosc_log.Transporter
	writer *_SysWriter
}

func (t *TransporterWindows) Reset(c interface{}, formatter eosc_log.Formatter) error {
	return nil
}

func (t *TransporterWindows) Close() error{
	return nil
}

//CreateTransporter 创建windows下的syslog，windows下不支持syslog，直接返回错误
func CreateTransporter(network, raddr string, level eosc_log.Level) (*TransporterWindows, error) {
	return nil, fmt.Errorf("can not create syslog transporterReset in windows system")
}
