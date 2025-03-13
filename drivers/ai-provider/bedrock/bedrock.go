package bedrock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/eolinker/eosc/log"

	"github.com/mitchellh/mapstructure"

	"github.com/aws/aws-sdk-go/aws/awserr"
	openai "github.com/sashabaranov/go-openai"

	"github.com/aws/aws-sdk-go/private/protocol/eventstream"

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
	httpContext.Response().AppendStreamFunc(c.streamHandler)
	ctx.SetLabel("response-content-type", "text/event-stream")
	return nil
}

func (c *Convert) ResponseConvert(ctx eocontext.EoContext) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	if httpContext.Response().StatusCode() != 200 {
		return nil
	}
	body := httpContext.Response().GetBody()
	var origin BedrockResponse

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

func (c *Convert) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
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

// EventStreamToJSON 将 Amazon EventStream 格式的数据转换为 JSON 格式
func EventStreamToJSON(eventStreamData []byte) ([]StreamResponse, error) {
	// 创建一个结果数组
	var result []StreamResponse

	// 创建一个 EventStream 解码器
	decoder := eventstream.NewDecoder(bytes.NewReader(eventStreamData))

	// 循环读取所有事件
	for {

		// 读取下一个消息
		msg, err := decoder.Decode(nil)
		if err != nil {
			if err == io.EOF {
				break // 正常结束
			}

			// 处理 AWS 错误
			if awsErr, ok := err.(awserr.Error); ok {
				return nil, fmt.Errorf("AWS Error: %s - %s", awsErr.Code(), awsErr.Message())
			}

			return nil, fmt.Errorf("解析 EventStream 时出错: %v", err)
		}

		// 将消息转换为 map
		eventMap := make(map[string]interface{})

		// 处理消息头
		headers := make(map[string]interface{})
		for _, header := range msg.Headers {
			headers[header.Name] = header.Value
		}

		eventMap["headers"] = headers

		// 处理消息体
		if len(msg.Payload) > 0 {
			// 尝试将负载解析为 JSON
			var payload interface{}
			if err := json.Unmarshal(msg.Payload, &payload); err == nil {
				eventMap["payload"] = payload
			} else {
				// 如果不是有效的 JSON，则作为字符串处理
				eventMap["payload"] = string(msg.Payload)
			}
		}
		var streamResponse StreamResponse
		err = mapstructure.Decode(eventMap, &streamResponse)
		if err != nil {
			return nil, err
		}

		// 将事件添加到结果数组
		result = append(result, streamResponse)
	}
	return result, nil
}

// ExtractTextFromEventStream 从 EventStream 中提取文本内容
// 这个函数专门用于从 Bedrock 模型响应中提取生成的文本
func ExtractTextFromEventStream(eventStreamData []byte) (string, error) {
	var fullText string

	decoder := eventstream.NewDecoder(bytes.NewReader(eventStreamData))

	for {
		msg, err := decoder.Decode(nil)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		// 解析消息负载
		var response map[string]interface{}
		if err := json.Unmarshal(msg.Payload, &response); err != nil {
			continue // 跳过无法解析的消息
		}

		// 根据 Bedrock 的响应格式提取文本
		// 注意：具体的字段名可能需要根据使用的模型进行调整
		if completion, ok := response["completion"].(string); ok {
			fullText += completion
		} else if output, ok := response["output"].(map[string]interface{}); ok {
			if text, ok := output["text"].(string); ok {
				fullText += text
			}
		}
	}

	return fullText, nil
}
