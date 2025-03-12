package bedrock

import (
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type ClientRequest struct {
	Messages        []*Message       `json:"message,omitempty"`
	System          *Content         `json:"system,omitempty"`
	InferenceConfig *InferenceConfig `json:"inferenceConfig,omitempty"`
}

type Message struct {
	Role    string     `json:"role"`
	Content []*Content `json:"content"`
}

type Content struct {
	Text string `json:"text"`
}

type InferenceConfig struct {
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"topP"`
}

// BedrockResponse 代表 Amazon Bedrock 的 JSON 响应格式
type BedrockResponse struct {
	Metrics struct {
		LatencyMs int `json:"latencyMs"`
	} `json:"metrics"`
	Output struct {
		Message struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Role string `json:"role"`
		} `json:"message"`
	} `json:"output"`
	StopReason string `json:"stopReason"`
	Usage      struct {
		InputTokens  int `json:"inputTokens"`
		OutputTokens int `json:"outputTokens"`
		TotalTokens  int `json:"totalTokens"`
	} `json:"usage"`
}

// ConvertBedrockToOpenAI 通用转换方法
func ConvertBedrockToOpenAI(requestId string, model string, bedrockResp BedrockResponse, isStream bool) openai.ChatCompletionResponse {
	// 提取文本内容
	textContent := ""
	if len(bedrockResp.Output.Message.Content) > 0 {
		textContent = bedrockResp.Output.Message.Content[0].Text
	}

	stopReason := openai.FinishReasonStop
	//end_turn | tool_use | max_tokens | stop_sequence | guardrail_intervened | content_filtered
	switch bedrockResp.StopReason {
	case "max_tokens":
		stopReason = openai.FinishReasonLength
	case "content_filtered":
		stopReason = openai.FinishReasonContentFilter
	}
	oj := "chat.completion"
	if isStream {
		oj = "chat.completion.chunk"
	}
	return openai.ChatCompletionResponse{
		ID:      requestId,
		Object:  oj,
		Created: time.Now().Unix(), // 这里可以替换为实际时间戳
		Model:   model,
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: textContent,
				},
				FinishReason: stopReason,
			},
		},
		Usage: openai.Usage{
			PromptTokens:     bedrockResp.Usage.InputTokens,
			CompletionTokens: bedrockResp.Usage.OutputTokens,
			TotalTokens:      bedrockResp.Usage.TotalTokens,
		},
	}
}

type StreamResponse struct {
	Header  Header  `json:"headers" mapstructure:"headers"`
	Payload Payload `json:"payload" mapstructure:"payload"`
}

type Header struct {
	ContentType string `json:":content-type" mapstructure:":content-type"`
	EventType   string `json:":event-type" mapstructure:":event-type"`
	MessageType string `json:":message-type" mapstructure:":message-type"`
}

type Payload struct {
	ContentBlockIndex int    `json:"contentBlockIndex" mapstructure:"contentBlockIndex"`
	Delta             *Delta `json:"delta" mapstructure:"delta"`
	P                 string `json:"p" mapstructure:"p"`
	StopReason        string `json:"stopReason,omitempty" mapstructure:"stopReason"`
	Metrics           struct {
		LatencyMs int `json:"latencyMs" mapstructure:"latencyMs"`
	} `json:"metrics,omitempty" mapstructure:"metrics"`
	Usage *Usage `json:"usage,omitempty" mapstructure:"usage"`
}

type Usage struct {
	InputTokens  int `json:"inputTokens" mapstructure:"inputTokens"`
	OutputTokens int `json:"outputTokens" mapstructure:"outputTokens"`
	TotalTokens  int `json:"totalTokens" mapstructure:"totalTokens"`
}

type Delta struct {
	Text string `json:"text" mapstructure:"text"`
}
