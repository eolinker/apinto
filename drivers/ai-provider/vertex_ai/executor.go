package vertex_ai

import (
	"context"
	"embed"
	"fmt"

	dns "google.golang.org/api/dns/v2"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

var (
	//go:embed vertex_ai.yaml
	providerContent []byte
	//go:embed *
	providerDir  embed.FS
	modelConvert = make(map[string]ai_convert.IChildConverter)

	_      ai_convert.IConverterDriver = (*executor)(nil)
	scopes                             = []string{
		dns.CloudPlatformReadOnlyScope,
		dns.CloudPlatformScope,
	}
)

func init() {
	models, err := ai_convert.LoadModels(providerContent, providerDir)
	if err != nil {
		panic(err)
	}
	for key, value := range models {
		if value.ModelProperties != nil {
			if f, ok := modelModes[value.ModelProperties.Mode]; ok {
				modelConvert[key] = f(value.Model)
			}
		}
	}
}

type executor struct {
	drivers.WorkerBase
	ai_convert.IConverterDriver
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
	d, err := newConverterDriver(conf)
	if err != nil {
		return err
	}
	e.IConverterDriver = d
	return nil
}

func (e *executor) Stop() error {
	e.IConverterDriver = nil
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return ai_convert.CheckKeySourceSkill(skill)
}

type ModelConfig struct {
	MaxOutputTokens  int     `json:"max_tokens"`
	Temperature      float64 `json:"temperature"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	TopP             float64 `json:"top_p"`
	TopK             int     `json:"top_k"`
}

func newToken(ctx context.Context, data []byte) (*oauth2.Token, error) {
	cfg, err := google.JWTConfigFromJSON(data, scopes...)
	if err != nil {
		return nil, err
	}
	return cfg.TokenSource(ctx).Token()
}
