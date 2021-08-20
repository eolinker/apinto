package httplog_transporter

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku/log/httplog"

	eosc_log "github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/log"
)

type Transporter struct {
	*eosc_log.Transporter
	writer *_HttpWriter
}

func (t *Transporter) Reset(c interface{}, formatter eosc_log.Formatter) error {
	conf, ok := c.(*httplog.Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*httplog.Config)(nil)), eosc.TypeNameOf(c))
	}

	t.Transporter.SetFormatter(formatter)
	return t.reset(conf)
}

func (t *Transporter) Close() error {
	t.Transporter.Close()
	return t.writer.Close()
}

func (t *Transporter) reset(c *httplog.Config) error {
	t.SetOutput(t.writer)
	t.SetLevel(c.Level)

	t.writer.reset(c)
	t.Transporter.SetOutput(t.writer)
	return nil
}

func CreateTransporter(conf *httplog.Config, formatter eosc_log.Formatter) (log.TransporterReset, error) {

	httpWriter := newHttpWriter()

	transport := &Transporter{
		Transporter: eosc_log.NewTransport(httpWriter, conf.Level, formatter),
		writer:      httpWriter,
	}

	e := transport.Reset(conf, formatter)
	if e != nil {
		return nil, e
	}
	return transport, nil
}
