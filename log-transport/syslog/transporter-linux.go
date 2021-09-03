//+build !windows

package syslog

import (
	"fmt"
	"github.com/eolinker/eosc"
	eosc_log "github.com/eolinker/eosc/log"
)

//Transporter syslog-Transporter结构
type Transporter struct {
	*eosc_log.Transporter
	writer *_SysWriter
}

//Reset 重置配置
func (t *Transporter) Reset(c interface{}, formatter eosc_log.Formatter) error {
	conf, ok := c.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(c))
	}

	t.Transporter.SetFormatter(formatter)
	return t.reset(conf)
}

func (t *Transporter) reset(c *Config) error {
	sysWriter, err := newSysWriter(c.Network, c.RAddr, c.Level, "")
	if err != nil {
		return err
	}
	err = t.writer.Close()
	if err != nil {
		return err
	}

	t.writer = sysWriter
	t.SetOutput(sysWriter)
	t.SetLevel(c.Level)

	return nil
}

//Close 关闭
func (t *Transporter) Close() error {
	t.Transporter.Close()
	return t.writer.Close()
}

//CreateTransporter 创建syslog-Transporter
func CreateTransporter(network, raddr string, level eosc_log.Level) (*Transporter, error) {

	sysWriter, err := newSysWriter(network, raddr, level, "")
	if err != nil {
		return nil, err
	}

	transport := &Transporter{
		Transporter: eosc_log.NewTransport(sysWriter, level),
		writer:      sysWriter,
	}

	return transport, nil
}
