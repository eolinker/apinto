package stdlog

import (
	"fmt"
	"os"

	log_transport "github.com/eolinker/goku/log-transport"

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
func CreateTransporter(conf *Config, formatter eosc_log.Formatter) (log_transport.TransporterReset, error) {

	transport := &Transporter{
		Transporter: eosc_log.NewTransport(os.Stdout, conf.Level, formatter),
		writer:      os.Stdout,
	}

	transport.SetLevel(conf.Level)

	return transport, nil
}
