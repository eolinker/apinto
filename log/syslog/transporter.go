//+build !windows

package syslog

import (
	"fmt"
	"github.com/eolinker/eosc"
	eosc_log "github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/log"
)

type Transporter struct {
	*eosc_log.Transporter
	writer *_SysWriter
}

func (t *Transporter) Reset(c interface{}, formatter eosc_log.Formatter) error {
	conf, ok := c.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(c))
	}

	t.Transporter.SetFormatter(formatter)
	return t.reset(conf)
}

func (t *Transporter) reset(c *Config) error {
	t.SetOutput(t.writer)
	t.SetLevel(c.Level)

	return nil
}

func createTransporter(conf *Config, formatter eosc_log.Formatter) (log.TransporterReset, error) {

	sysWriter, err := newSysWriter(conf.Network, conf.RAddr, conf.Level, "")
	if err != nil {
		return nil, err
	}

	transport := &Transporter{
		Transporter: eosc_log.NewTransport(sysWriter, conf.Level, formatter),
		writer:      sysWriter,
	}

	e := transport.Reset(conf, formatter)
	if e != nil {
		return nil, e
	}
	return transport, nil
}
