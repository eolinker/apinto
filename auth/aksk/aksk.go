package aksk

import (
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/auth"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
)

//supportTypes 当前驱动支持的authorization type值
var supportTypes = []string{
	"ak/sk",
	"aksk",
}

type aksk struct {
	id         string
	name       string
	labels     map[string]string
	akskConfig map[string]AKSKConfig
}

func (a *aksk) Id() string {
	return a.id
}

func (a *aksk) Start() error {
	return nil
}

func (a *aksk) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	config, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf))
	}

	for _, c := range config.akskConfig {
		a.akskConfig[c.AK] = c
	}

	return nil
}

func (a *aksk) Stop() error {
	return nil
}

func (a *aksk) CheckSkill(skill string) bool {
	return auth.CheckSkill(skill)
}

func (a *aksk) Auth(context *http_context.Context) error {
	return nil
}
