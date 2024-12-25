// Open AI单元测试
package openAI

import (
	"net/url"
	"os"
	"testing"
	"time"

	http_context "github.com/eolinker/apinto/node/http-context"

	"github.com/eolinker/apinto/convert"

	"github.com/valyala/fasthttp"
)

var defaultConfig = `{
  "frequency_penalty": "",
  "max_tokens": 512,
  "presence_penalty": "",
  "response_format": "",
  "temperature": "",
  "top_p": ""
}`

type tmpConfig struct {
	id          string
	name        string
	cfg         *Config
	model       string
	modelConfig string
	requestBody []tmpRequest
}

type tmpRequest struct {
	body []byte
}

func TestSentTo(t *testing.T) {
	apikey := os.Getenv("OPENAI_API_KEY")
	cfg := &Config{
		APIKey: apikey,
		Base:   "https://api.openai-proxy.com",
	}
	worker, err := Create("openai", "openai", cfg, nil)

	if err != nil {
		t.Fatal(err)
	}
	handler, ok := worker.(convert.IConverterDriver)
	if !ok {
		return
	}
	model := "gpt-3.5-turbo"
	fn, has := handler.GetModel(model)
	if !has {
		t.Fatal("model not found")
	}

	extender, err := fn(defaultConfig)
	if err != nil {
		t.Fatalf("generate config error: %v", err)
	}
	body := []byte(`{
  "messages": [
    {
      "content": "Hello, how can I help you?",
      "role": "assistant"
    }
  ],
  "variables": {
    "source_lang": "",
    "target_lang": "",
    "text": ""
  }
}`)
	ctx := http_context.NewContext(genRequestCtx("/xxx/xxx", nil, nil, body), 8099)

	converter, has := handler.GetConverter(model)
	if !has {
		t.Fatal("converter not found")
	}
	err = converter.RequestConvert(ctx, extender)
	if err != nil {
		t.Fatal(err)
	}
	balance := ctx.GetBalance()
	node, _, err := balance.Select(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = ctx.SendTo(balance.Scheme(), node, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	err = converter.ResponseConvert(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func genRequestCtx(rawURL string, headers map[string]string, query url.Values, body []byte) *fasthttp.RequestCtx {
	req := fasthttp.AcquireRequest()
	u := fasthttp.AcquireURI()
	uri, _ := url.Parse(rawURL)
	u.SetPath(uri.Path)
	u.SetScheme(uri.Scheme)
	u.SetHost(uri.Host)
	u.SetQueryString(uri.RawQuery)
	req.SetURI(u)
	req.SetHostBytes(u.Host())
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.SetHostBytes(u.Host())
	req.SetBody(body)
	return &fasthttp.RequestCtx{
		Request:  *req,
		Response: fasthttp.Response{},
	}
}
