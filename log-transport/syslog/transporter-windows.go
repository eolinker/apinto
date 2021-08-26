//+build windows
package syslog

import (
	"fmt"

	eosc_log "github.com/eolinker/eosc/log"
)

func CreateTransporter(conf *Config, formatter eosc_log.Formatter) (log_transport.TransporterReset, error) {
	return nil, fmt.Errorf("can not create syslog transporterReset in windows system")
}
