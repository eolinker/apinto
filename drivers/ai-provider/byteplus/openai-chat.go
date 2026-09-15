package byteplus

import (
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeOpenAIChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.BaseUrl, mt, time.Minute*10, nil, errorCallback)
	})

	driverCreate.Set(ai_convert.ModelTypeChat, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return ai_convert.NewOpenAIChat(provider, c.APIKey, c.BaseUrl, mt, time.Minute*10, nil, errorCallback)
	})
}

