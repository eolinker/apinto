package filelog

import (
	"fmt"
	"github.com/eolinker/eosc"
	eosc_log "github.com/eolinker/eosc/log"
	transporterManager "github.com/eolinker/eosc/log/transporter-manager"
	"github.com/eolinker/goku/log"
	logFormatter "github.com/eolinker/goku/log/common/log-formatter"
)

type filelog struct {
	id                 string
	name               string
	config             *Config
	formatterName      string
	transporterManager transporterManager.ITransporterManager
}

func (f *filelog) Id() string {
	return f.id
}

func (f *filelog) Transport(entry *eosc_log.Entry) error {
	return nil
}

func (f *filelog) Start() error {

	formatter := logFormatter.CreateFormatter(driverName, f.formatterName)
	transporterReset, err := createTransporter(f.config, formatter)
	if err != nil {
		return err
	}

	return f.transporterManager.Set(f.id, transporterReset)
}

func (f *filelog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, ok := conf.(*DriverConfig)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*DriverConfig)(nil)), eosc.TypeNameOf(conf))
	}

	c, err := ToConfig(config)
	if err != nil {
		return err
	}
	f.config = c
	f.formatterName = config.FormatterName

	formatter := logFormatter.CreateFormatter(driverName, f.formatterName)
	transporter, err := createTransporter(f.config, formatter)
	if err != nil {
		return err
	}

	return f.transporterManager.Set(f.id, transporter)
}

func (f *filelog) Stop() error {
	return f.transporterManager.Del(f.id)
}

func (f *filelog) CheckSkill(skill string) bool {
	return log.CheckSkill(skill)
}
