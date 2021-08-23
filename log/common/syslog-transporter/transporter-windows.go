//+build windows
package syslog_transporter

import (
	"fmt"
	eosc_log "github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/log"
)

func CreateTransporter(conf *Config, formatter eosc_log.Formatter) (log.TransporterReset, error) {
	return nil, fmt.Errorf("can not create syslog transporterReset in windows system")
}
