package httplog

import (
	"fmt"

	"github.com/eolinker/eosc"
	eosc_log "github.com/eolinker/eosc/log"
)

//Transporter httplog-Transporter结构
type Transporter struct {
	*eosc_log.Transporter
	writer *_HttpWriter
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

//Close 关闭
func (t *Transporter) Close() error {
	t.Transporter.Close()
	return t.writer.Close()
}

func (t *Transporter) reset(c *Config) error {
	t.SetLevel(c.Level)

	t.writer.reset(c)
	t.Transporter.SetOutput(t.writer)
	return nil
}

//CreateTransporter 创建httplog-Transporter
func CreateTransporter(level eosc_log.Level) *Transporter {
	httpWriter := newHTTPWriter()

 	return &Transporter{
		Transporter: eosc_log.NewTransport(httpWriter, level),
		writer:      httpWriter,
	}
}
