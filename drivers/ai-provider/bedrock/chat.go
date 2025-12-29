package bedrock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"strings"
	"time"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeChat, func(conf *Config) (ai_convert.IConverterDriver, error) {
		return &Chat{
			signer: v4.NewSigner(credentials.NewStaticCredentials(conf.AccessKey, conf.SecretKey, "")),
			region: conf.Region,
		}, nil
	})
}

type Chat struct {
	signer *v4.Signer
	region string
}

func (c *Chat) Provider() string {
	return provider
}

func (c *Chat) ModelType() ai_convert.ModelType {
	return ai_convert.ModelTypeChat
}

var (
	currentPath = "/model/%s/converse"
	streamPath  = "/model/%s/converse-stream"
)

func (c *Chat) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
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
	httpContext.Proxy().AppendStreamBodyHandle(c.streamHandler)
	ctx.SetLabel("response-content-type", "text/event-stream")
	return nil
}

func (c *Chat) ResponseConvert(ctx eocontext.EoContext) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	if httpContext.Response().StatusCode() != 200 {
		return nil
	}
	body := httpContext.Response().GetBody()
	var origin Response

	err = json.Unmarshal(body, &origin)
	if err != nil {
		return err
	}
	resp := ConvertBedrockToOpenAI(ctx.RequestId(), ai_convert.GetAIModel(ctx), origin, false)

	body, err = json.Marshal(resp)
	if err != nil {
		return err
	}

	httpContext.Response().SetBody(body)
	return nil
}

func (c *Chat) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
	// 创建一个缓冲区来存储转换后的SSE格式数据
	var sseBuffer bytes.Buffer

	// 生成一个唯一的请求ID
	requestID := ctx.RequestId()
	model := ai_convert.GetAIModel(ctx)
	response, err := EventStreamToJSON(p)
	if err != nil {
		log.Errorf("event stream to json error: %v", err)
		return p, nil
	}

	for _, r := range response {
		switch r.Header.EventType {
		case "contentBlockDelta":
			{
				if r.Payload.Delta != nil {
					data := openai.ChatCompletionStreamResponse{
						ID:      requestID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   model,
						Choices: []openai.ChatCompletionStreamChoice{
							{
								Index: 0,
								Delta: openai.ChatCompletionStreamChoiceDelta{
									Content: r.Payload.Delta.Text,
									Role:    "assistant",
								},
							},
						},
					}
					content, _ := json.Marshal(data)
					sseBuffer.WriteString(fmt.Sprintf("data: %s\n\n", string(content)))
				}
			}
		case "messageStop":
			{
				usage := new(openai.Usage)
				if r.Payload.Usage != nil {
					usage.PromptTokens = r.Payload.Usage.InputTokens
					usage.CompletionTokens = r.Payload.Usage.OutputTokens
					usage.TotalTokens = r.Payload.Usage.TotalTokens
					ai_convert.SetAIModelInputToken(ctx, r.Payload.Usage.InputTokens)
					ai_convert.SetAIModelOutputToken(ctx, r.Payload.Usage.OutputTokens)
					ai_convert.SetAIModelTotalToken(ctx, r.Payload.Usage.TotalTokens)
				}
				stopReason := openai.FinishReasonStop
				//end_turn | tool_use | max_tokens | stop_sequence | guardrail_intervened | content_filtered
				switch r.Payload.StopReason {
				case "max_tokens":
					stopReason = openai.FinishReasonLength
				case "content_filtered":
					stopReason = openai.FinishReasonContentFilter
				}
				data := openai.ChatCompletionStreamResponse{
					ID:      ctx.RequestId(),
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   ai_convert.GetAIModel(ctx),
					Choices: []openai.ChatCompletionStreamChoice{
						{
							Index:        0,
							FinishReason: stopReason,
						},
					},
					Usage: usage,
				}
				content, _ := json.Marshal(data)
				sseBuffer.WriteString(fmt.Sprintf("data: %s\n\n", string(content)))
				sseBuffer.WriteString("data: [DONE]\n\n")
				return sseBuffer.Bytes(), nil
			}
		}
	}

	// 返回转换后的SSE格式数据
	return sseBuffer.Bytes(), nil
}
