package bedrock

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/apinto/convert"
	"github.com/eolinker/eosc"
)

var (
	//go:embed bedrock.yaml
	providerContent []byte
	//go:embed *
	providerDir  embed.FS
	modelConvert = make(map[string]convert.IConverter)

	_ convert.IConverterDriver = (*executor)(nil)
)

func init() {
	models, err := convert.LoadModels(providerContent, providerDir)
	if err != nil {
		panic(err)
	}
	for key, value := range models {
		if value.ModelProperties != nil {
			if v, ok := modelModes[value.ModelProperties.Mode]; ok {
				modelConvert[key] = v(value.Model)
			}
		}
	}
}

type executor struct {
	drivers.WorkerBase
	convert.IConverterDriver
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
	return convert.CheckSkill(skill)
}

type ModelConfig struct {
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
}

func signRequest(signer *v4.Signer, region string, model string, headers http.Header, body string) (http.Header, error) {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/converse", region, model), nil)
	if err != nil {
		return nil, err
	}
	request.Header = headers.Clone()

	_, err = signer.Sign(request, strings.NewReader(body), "bedrock", region, time.Now())
	if err != nil {
		return nil, err
	}
	return request.Header, nil

}
