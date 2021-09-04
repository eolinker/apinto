package filelog

import (
	"fmt"
	"time"

	"github.com/eolinker/eosc"
	eosc_log "github.com/eolinker/eosc/log"
)

//Transporter filelog-Transporter结构
type Transporter struct {
	*eosc_log.Transporter
	writer *FileWriterByPeriod
}

//Close 关闭
func (t *Transporter) Close() error {
	t.writer.Close()
	return nil
}

//Reset 重置配置
func (t *Transporter) Reset(c interface{}, f eosc_log.Formatter) error {
	conf, ok := c.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(c))
	}

	t.Transporter.SetFormatter(f)
	return t.reset(conf)
}

func (t *Transporter) reset(c *Config) error {
	t.SetOutput(t.writer)
	t.SetLevel(c.Level)

	t.writer.Set(
		c.Dir,
		c.File,
		c.Period,
		time.Duration(c.Expire)*time.Hour*24,
	)
	t.writer.Open()
	return nil
}

//CreateTransporter 创建filelog-Transporter
func CreateTransporter(level eosc_log.Level) *Transporter {

	fileWriterByPeriod := NewFileWriteByPeriod()

	return &Transporter{
		Transporter: eosc_log.NewTransport(fileWriterByPeriod, level),
		writer:      fileWriterByPeriod,
	}
}
