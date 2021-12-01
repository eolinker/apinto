package stdlog

import (
	"fmt"
	"os"

	"github.com/eolinker/eosc"
	eosc_log "github.com/eolinker/eosc/log"
)

//Transporter stdlog-Transporter结构
type Transporter struct {
	*eosc_log.Transporter
	writer *os.File
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
	t.SetLevel(c.Level)

	return nil
}

//CreateTransporter 创建stdlog-Transporter
func CreateTransporter(level eosc_log.Level) *Transporter {

	return &Transporter{
		Transporter: eosc_log.NewTransport(os.Stdout, level),
		writer:      os.Stdout,
	}
}
