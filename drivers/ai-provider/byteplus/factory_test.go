package byteplus

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"testing"
)

func TestCreate(t *testing.T) {
	cfg := `{
		"api_key":"xxxx",
"base_url":"https://ark.ap-southeast.bytepluses.com/api/v3"
	}`
	c, err := Create(cfg)
	if err != nil {
		panic(err)
	}
	if c == nil {
		panic("converter is nil")
	}
	ai_convert.GetConverterCreateFunc(provider)
}
