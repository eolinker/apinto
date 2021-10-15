package log_transport

import "github.com/eolinker/eosc/log"

//TransporterReset 实现了可重置配置的EntryTransporter
type TransporterReset interface {
	log.EntryTransporter
	Reset(config interface{}, formatter log.Formatter) error
}
