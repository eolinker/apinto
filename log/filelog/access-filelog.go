package filelog

import (
	"fmt"
	"github.com/eolinker/eosc"
	eosc_log "github.com/eolinker/eosc/log"
	"github.com/eolinker/goku/log"
)

type filelogAccess struct {
	id                 string
	name               string
	config             *Config
	formatterName      string
	fields             []string
	transporterManager transporterManager.ITransporterManager
}

func (f *filelogAccess) Id() string {
	return f.id
}

func (f *filelogAccess) Transport(entry *eosc_log.Entry) error {
	return nil
}

func (f *filelogAccess) Start() error {
	//TODO 组装formatter

	transporter, err := createTransporter(f.config)
	if err != nil {
		return err
	}

	return f.transporterManager.Set(f.id, transporter)
}

func (f *filelogAccess) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, ok := conf.(*DriverConfigAccess)

	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*DriverConfigAccess)(nil)), eosc.TypeNameOf(conf))
	}

	c, err := ToConfigAccess(config)
	if err != nil {
		return err
	}
	f.config = c
	f.formatterName = config.FormatterName
	f.fields = config.Fields

	//TODO 组装formatter
	transporter, err := createTransporter(f.config, nil)
	if err != nil {
		return err
	}

	return f.transporterManager.Set(f.id, transporter)
}

func (f *filelogAccess) Stop() error {
	return f.transporterManager.Del(f.id)
}

func (f *filelogAccess) CheckSkill(skill string) bool {
	return log.CheckSkill(skill)
}
