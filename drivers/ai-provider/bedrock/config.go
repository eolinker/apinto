package bedrock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"

	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/eolinker/eosc/common/bean"
	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/eosc/eocontext"

	ai_convert "github.com/eolinker/apinto/ai-convert"

	"github.com/eolinker/eosc"
)

var (
	accessConfigManager ai_convert.IModelAccessConfigManager
)

func init() {
	bean.Autowired(&accessConfigManager)
	ai_convert.RegisterConverterCreateFunc("bedrock", Create)
}

type Config struct {
	AccessKey string `json:"aws_access_key_id"`
	SecretKey string `json:"aws_secret_access_key"`
	Region    string `json:"aws_region"`
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	if conf.AccessKey == "" {
		return nil, fmt.Errorf("aws_access_key_id is required")
	}
	if conf.SecretKey == "" {
		return nil, fmt.Errorf("aws_secret_access_key is required")
	}
	return NewConvert(conf.AccessKey, conf.SecretKey, conf.Region), nil
}

type Convert struct {
	signer *v4.Signer
	region string
}

func NewConvert(ak string, sk string, region string) *Convert {
	return &Convert{
		signer: v4.NewSigner(credentials.NewStaticCredentials(ak, sk, "")),
		region: region,
	}
}

var (
	currentPath = "/model/%s/converse"
	streamPath  = "/model/%s/converse-stream"
)

func (c *Convert) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	provider := ai_convert.GetAIProvider(ctx)
	model := ai_convert.GetAIModel(ctx)
	modelCfg, has := accessConfigManager.Get(fmt.Sprintf("%s$%s", provider, model))
	region := ""
	if has {
		model = modelCfg.Config()["model"]
		region = modelCfg.Config()["region"]
	}
	if region == "" {
		region = c.region
	}
	base := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com", region)

	balanceHandler, err := ai_convert.NewBalanceHandler("", base, 0)
	if err != nil {
		return err
	}
	ctx.SetBalance(balanceHandler)
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	chatRequest := eosc.NewBase[ai_convert.Request](extender)
	err = json.Unmarshal(body, chatRequest)
	if err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}
	messages := make([]Message, 0, len(chatRequest.Config.Messages))
	systemMessage := make([]*Content, 0)
	for _, m := range chatRequest.Config.Messages {
		if m.Role == "system" {
			systemMessage = append(systemMessage, &Content{Text: m.Content})
		} else {
			messages = append(messages, Message{
				Role:    m.Role,
				Content: []*Content{{Text: m.Content}},
			})
		}
	}
	chatRequest.SetAppend("messages", messages)
	chatRequest.SetAppend("system", systemMessage)
	path := fmt.Sprintf(currentPath, model)
	if chatRequest.Config.Stream {
		path = fmt.Sprintf(streamPath, model)
	}
	uri := fmt.Sprintf("%s%s", base, path)
	httpContext.Proxy().URI().SetPath(path)

	body, _ = json.Marshal(chatRequest)
	httpContext.Proxy().Body().SetRaw("application/json", body)
	headers, err := signRequest(c.signer, region, uri, http.Header{}, string(body))
	if err != nil {
		return err
	}
	for k, v := range headers {
		httpContext.Proxy().Header().SetHeader(k, strings.Join(v, ";"))
	}
	httpContext.Proxy().Body().SetRaw("application/json", body)
	return nil
}

func (c *Convert) ResponseConvert(ctx eocontext.EoContext) error {
	//httpContext, err := http_service.Assert(ctx)
	//if err != nil {
	//	return err
	//}
	//if httpContext.Response().StatusCode() != 200 {
	//	return nil
	//}
	//body := httpContext.Response().GetBody()
	//data := eosc.NewBase[Response](nil)
	//err = json.Unmarshal(body, data)
	//if err != nil {
	//	return err
	//}
	//responseBody := &ai_convert.Response{}
	//
	//body, err = json.Marshal(responseBody)
	//if err != nil {
	//	return err
	//}
	//httpContext.Response().AppendStreamFunc(c.streamHandler)
	//httpContext.Response().SetBody(body)
	return nil
}

func (c *Convert) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
	//// 对响应数据进行划分
	//inputToken := GetAIModelInputToken(ctx)
	//outputToken := 0
	//totalToken := inputToken
	//scanner := bufio.NewScanner(bytes.NewReader(p))
	//// Check the content encoding and convert to UTF-8 if necessary.
	//encoding := ctx.Response().Headers().Get("content-encoding")
	//for scanner.Scan() {
	//	line := scanner.Text()
	//	if encoding != "utf-8" && encoding != "" {
	//		tmp, err := encoder.ToUTF8(encoding, []byte(line))
	//		if err != nil {
	//			log.Errorf("convert to utf-8 error: %v, line: %s", err, line)
	//			return p, nil
	//		}
	//		if ctx.Response().StatusCode() != 200 || (o.checkErr != nil && !o.checkErr(ctx, tmp)) {
	//			if o.errorCallback != nil {
	//				o.errorCallback(ctx, tmp)
	//			}
	//			return p, nil
	//		}
	//		line = string(tmp)
	//	}
	//	line = strings.TrimPrefix(line, "data:")
	//	if line == "" || strings.Trim(line, " ") == "[DONE]" {
	//		return p, nil
	//	}
	//	var resp openai.ChatCompletionResponse
	//	err := json.Unmarshal([]byte(line), &resp)
	//	if err != nil {
	//		return p, nil
	//	}
	//	if len(resp.Choices) > 0 {
	//		outputToken += getTokens(resp.Choices[0].Message.Content)
	//		totalToken += outputToken
	//	}
	//}
	//if err := scanner.Err(); err != nil {
	//	log.Errorf("scan error: %v", err)
	//	return p, nil
	//}
	//
	//SetAIModelInputToken(ctx, inputToken)
	//SetAIModelOutputToken(ctx, outputToken)
	//SetAIModelTotalToken(ctx, totalToken)
	return p, nil
}

func signRequest(signer *v4.Signer, region string, uri string, headers http.Header, body string) (http.Header, error) {
	request, err := http.NewRequest(http.MethodPost, uri, nil)
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
