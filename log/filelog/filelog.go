package filelog

import (
	"fmt"
	"github.com/eolinker/eosc"
	transporterManager "github.com/eolinker/eosc/log/transporter-manager"
	"github.com/eolinker/goku/log"
	"github.com/eolinker/goku/log/filelog/filelog-transport"
)

type filelog struct {
	id                 string
	name               string
	config             *filelog_transport.Config
	formatterName      string
	transporterReset   log.TransporterReset
	transporterManager transporterManager.ITransporterManager
}

func (f *filelog) Id() string {
	return f.id
}

func (f *filelog) Start() error {

	formatter, err := filelog_transport.CreateFormatter(f.formatterName)
	if err != nil {
		return err
	}

	transporterReset, err := filelog_transport.CreateTransporter(f.config, formatter)
	if err != nil {
		return err
	}

	f.transporterReset = transporterReset
	return f.transporterManager.Set(f.id, transporterReset)
}

func (f *filelog) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, ok := conf.(*DriverConfig)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*DriverConfig)(nil)), eosc.TypeNameOf(conf))
	}

	c, err := toConfig(config)
	if err != nil {
		return err
	}
	f.config = c
	f.formatterName = config.FormatterName

	formatter, err := filelog_transport.CreateFormatter(f.formatterName)
	if err != nil {
		return err
	}

	err = f.transporterReset.Reset(c, formatter)
	if err != nil {
		return err
	}

	return nil
}

func (f *filelog) Stop() error {
	f.transporterReset.Close()
	return f.transporterManager.Del(f.id)
}

func (f *filelog) CheckSkill(skill string) bool {
	return log.CheckSkill(skill)
}
