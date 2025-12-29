package vertex_ai

import (
	"encoding/json"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/common/bean"

	dns "google.golang.org/api/dns/v2"
)

const (
	provider = "vertex_ai"
)

var (
	accessConfigManager ai_convert.IModelAccessConfigManager
	scopes              = []string{
		dns.CloudPlatformReadOnlyScope,
		dns.CloudPlatformScope,
	}
	driverCreate = eosc.BuildUntyped[ai_convert.ModelType, ai_convert.IConvertDriverCreateFunc[Config]]()
)

func init() {
	ai_convert.RegisterConverterCreateFunc(provider, Create)
	bean.Autowired(&accessConfigManager)
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}

	return ai_convert.NewConverter(provider, &conf, driverCreate.List())
}
