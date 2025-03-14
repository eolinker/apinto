package ai_convert

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/eolinker/eosc/log"

	http_service "github.com/eolinker/eosc/eocontext/http-context"

	"github.com/eolinker/apinto/encoder"
	"github.com/eolinker/eosc"
	eoscContext "github.com/eolinker/eosc/eocontext"
	openai "github.com/sashabaranov/go-openai"
)

const (
	OpenAIChatCompletePath = "/chat/completions"
)

type CheckError func(ctx http_service.IHttpContext, body []byte) bool

type OpenAIConvert struct {
	apikey         string
	path           string
	checkErr       CheckError
	errorCallback  func(ctx http_service.IHttpContext, body []byte)
	balanceHandler eoscContext.BalanceHandler
}

func NewOpenAIConvert(apikey string, baseUrl string, timeout time.Duration, checkErr CheckError, errorCallback func(ctx http_service.IHttpContext, body []byte)) (*OpenAIConvert, error) {
	c := &OpenAIConvert{
		apikey:        apikey,
		checkErr:      checkErr,
		errorCallback: errorCallback,
	}
	if baseUrl != "" {
		balanceHandler, err := NewBalanceHandler(apikey, baseUrl, timeout)
		if err != nil {
			return nil, err
		}
		c.balanceHandler = balanceHandler
		u, err := url.Parse(baseUrl)
		if err != nil {
			return nil, err
		}
		if strings.TrimSuffix(u.Path, "/") == "" {
			u.Path = "/v1"
		}
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), OpenAIChatCompletePath)
	} else {
		c.path = fmt.Sprintf("/v1%s", OpenAIChatCompletePath)
	}
	return c, nil
}

func (o *OpenAIConvert) RequestConvert(ctx eoscContext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	var promptToken int
	chatRequest := eosc.NewBase[Request](extender)
	err = json.Unmarshal(body, chatRequest)
	if err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}
	if chatRequest.Config.Model == "" {
		chatRequest.Config.Model = GetAIModel(ctx)
	}
	for _, msg := range chatRequest.Config.Messages {
		promptToken += getTokens(msg.Content)
	}
	SetAIModelInputToken(httpContext, promptToken)
	httpContext.Proxy().AppendStreamBodyHandle(o.streamHandler)
	if o.apikey != "" {
		httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+o.apikey)
	}

	httpContext.Proxy().URI().SetPath(o.path)
	body, _ = json.Marshal(chatRequest)
	httpContext.Proxy().Body().SetRaw("application/json", body)
	if o.balanceHandler != nil {
		ctx.SetBalance(o.balanceHandler)
	}
	httpContext.Proxy().AppendBodyFinish(o.bodyFinish)

	return nil
}

func ResponseConvert(ctx eoscContext.EoContext, checkErr CheckError, errorCallback func(ctx http_service.IHttpContext, body []byte)) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	body := httpContext.Response().GetBody()
	// Check the content encoding and convert to UTF-8 if necessary.
	encoding := httpContext.Response().Headers().Get("content-encoding")
	if encoding != "utf-8" && encoding != "" {
		body, err = encoder.ToUTF8(encoding, body)
		if err != nil {
			return err
		}
	}

	if (checkErr != nil && !checkErr(httpContext, body)) || httpContext.Response().StatusCode() != 200 {
		if errorCallback != nil {
			errorCallback(httpContext, body)
		}
		status := GetAIStatus(ctx)
		if status == "" {
			status = StatusInvalid
		}
		SetAIProviderStatuses(httpContext, AIProviderStatus{
			Provider: GetAIProvider(ctx),
			Model:    GetAIModel(ctx),
			Key:      GetAIKey(ctx),
			Status:   status,
		})
		return nil
	}

	var resp openai.ChatCompletionResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		SetAIProviderStatuses(httpContext, AIProviderStatus{
			Provider: GetAIProvider(ctx),
			Model:    GetAIModel(ctx),
			Key:      GetAIKey(ctx),
			Status:   StatusInvalid,
		})
		log.Errorf("unmarshal body error: %v, body: %s", err, string(body))
		return err
	}

	SetAIModelInputToken(httpContext, resp.Usage.PromptTokens)
	SetAIModelOutputToken(httpContext, resp.Usage.CompletionTokens)
	SetAIModelTotalToken(httpContext, resp.Usage.TotalTokens)
	SetAIStatusNormal(ctx)
	SetAIProviderStatuses(httpContext, AIProviderStatus{
		Provider: GetAIProvider(ctx),
		Model:    GetAIModel(ctx),
		Key:      GetAIKey(ctx),
		Status:   GetAIStatus(ctx),
	})
	httpContext.Response().SetHeader("content-encoding", "utf-8")
	httpContext.Response().SetBody(body)
	return nil
}

func (o *OpenAIConvert) ResponseConvert(ctx eoscContext.EoContext) error {
	return ResponseConvert(ctx, o.checkErr, o.errorCallback)
}

func (o *OpenAIConvert) bodyFinish(ctx http_service.IHttpContext) {
	body := ctx.Response().GetBody()
	defer func() {
		SetAIProviderStatuses(ctx, AIProviderStatus{
			Provider: GetAIProvider(ctx),
			Model:    GetAIModel(ctx),
			Key:      GetAIKey(ctx),
			Status:   GetAIStatus(ctx),
		})
	}()
	if o.checkErr != nil && !o.checkErr(ctx, body) {
		o.errorCallback(ctx, body)
		return
	}
	encoding := ctx.Response().Headers().Get("content-encoding")
	if encoding == "gzip" {
		tmp, err := encoder.ToUTF8(encoding, body)
		if err != nil {
			log.Errorf("convert to utf-8 error: %v, body: %s", err, string(body))
			return
		}
		var resp openai.ChatCompletionResponse
		err = json.Unmarshal(tmp, &resp)
		if err != nil {
			log.Errorf("unmarshal body error: %v, body: %s", err, string(tmp))
			return
		}
		SetAIModelInputToken(ctx, resp.Usage.PromptTokens)
		SetAIModelOutputToken(ctx, resp.Usage.CompletionTokens)
		SetAIModelTotalToken(ctx, resp.Usage.TotalTokens)
	}
	SetAIStatusNormal(ctx)
}

func (o *OpenAIConvert) streamHandler(ctx http_service.IHttpContext, p []byte) ([]byte, error) {
	encoding := ctx.Response().Headers().Get("content-encoding")
	if encoding == "gzip" {
		return p, nil
	}

	// 对响应数据进行划分
	inputToken := GetAIModelInputToken(ctx)
	outputToken := GetAIModelOutputToken(ctx)
	totalToken := inputToken
	defer func() {
		SetAIModelInputToken(ctx, inputToken)
		SetAIModelOutputToken(ctx, outputToken)
		SetAIModelTotalToken(ctx, totalToken)
	}()
	scanner := bufio.NewScanner(bytes.NewReader(p))
	for scanner.Scan() {
		line := scanner.Text()
		if encoding != "utf-8" && encoding != "" {
			tmp, err := encoder.ToUTF8(encoding, []byte(line))
			if err != nil {
				log.Errorf("convert to utf-8 error: %v, line: %s", err, line)
				return p, nil
			}
			line = string(tmp)
		}

		line = strings.TrimPrefix(line, "data:")
		if line == "" || strings.Trim(line, " ") == "[DONE]" {
			return p, nil
		}
		var resp openai.ChatCompletionStreamResponse
		err := json.Unmarshal([]byte(line), &resp)
		if err != nil {
			return p, nil
		}
		if len(resp.Choices) > 0 {
			outputToken += getTokens(resp.Choices[0].Delta.Content)
			totalToken += outputToken
		}
		if resp.Usage != nil {
			inputToken = resp.Usage.PromptTokens
			outputToken = resp.Usage.CompletionTokens
			totalToken = resp.Usage.TotalTokens
		}
	}

	return p, nil
}
