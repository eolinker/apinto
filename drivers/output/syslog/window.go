//go:build windows
// +build windows

package syslog

import (
	"fmt"
)

//CreateTransporter 创建windows下的syslog，windows下不支持syslog，直接返回错误
func CreateTransporter(conf *SysConfig) (*SysWriter, error) {
	return nil, fmt.Errorf("can not create syslog transporterReset in windows system")
}
