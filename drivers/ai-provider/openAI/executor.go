package openAI

import (
	"embed"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
)

var (
	//go:embed openai.yaml
	providerContent []byte
	//go:embed *
	providerDir  embed.FS
	modelConvert = make(map[string]convert.IConverter)

	_ convert.IConverterDriver = (*executor)(nil)
)

func init() {
	models, err := ai_provider.LoadModels(providerContent, providerDir)
	if err != nil {
		panic(err)
	}
	for key, value := range models {
		if value.ModelProperties != nil {
			if v, ok := modelModes[value.ModelProperties.Mode]; ok {
				modelConvert[key] = v
			}
		}
	}
}

type executor struct {
	drivers.WorkerBase
	apikey string
	eocontext.BalanceHandler
}

type Converter struct {
	balanceHandler eocontext.BalanceHandler
	converter      convert.IConverter
}

func (c *Converter) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}
	return c.converter.RequestConvert(ctx, extender)
}

func (c *Converter) ResponseConvert(ctx eocontext.EoContext) error {
	return c.converter.ResponseConvert(ctx)
}

func (e *executor) GetConverter(model string) (convert.IConverter, bool) {
	converter, ok := modelConvert[model]
	if !ok {
		return nil, false
	}

	return &Converter{balanceHandler: e.BalanceHandler, converter: converter}, true
}

func (e *executor) GetModel(model string) (convert.FGenerateConfig, bool) {
	if _, ok := modelConvert[model]; !ok {
		return nil, false
	}
	return func(cfg string) (map[string]interface{}, error) {
		return nil, nil
	}, true
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("invalid config")
	}

	return e.reset(cfg, workers)
}

func (e *executor) reset(conf *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	if conf.Base != "" {
		u, err := url.Parse(conf.Base)
		if err != nil {
			return err
		}
		hosts := strings.Split(u.Host, ":")
		ip := hosts[0]
		port := 80
		if u.Scheme == "https" {
			port = 443
		}
		if len(hosts) > 1 {
			port, _ = strconv.Atoi(hosts[1])
		}
		e.BalanceHandler = ai_provider.NewBalanceHandler(u.Scheme, 0, []eocontext.INode{ai_provider.NewBaseNode(e.Id(), ip, port)})
	} else {
		e.BalanceHandler = nil
	}

	e.apikey = conf.APIKey
	return nil
}

func (e *executor) Stop() error {
	e.BalanceHandler = nil
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return convert.CheckSkill(skill)
}
