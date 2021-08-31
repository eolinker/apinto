package log_transport

import "github.com/eolinker/eosc/log"

type TransporterReset interface {
	log.EntryTransporter
	Reset(config interface{}, formatter log.Formatter) error
}
