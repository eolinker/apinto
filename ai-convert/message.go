package ai_convert

import (
	tiktoken "github.com/pkoukk/tiktoken-go"
	openai "github.com/sashabaranov/go-openai"
)

// Request 定义客户端统一输入请求格式
type Request openai.ChatCompletionRequest

// Response 定义客户端统一输出响应格式
type Response struct {
	*openai.ChatCompletionResponse
	Error *Error `json:"error,omitempty"`
}

type ModelConfig struct {
	MaxTokens           int                                  `json:"max_tokens,omitempty"`
	MaxCompletionTokens int                                  `json:"max_completion_tokens,omitempty"`
	Temperature         float32                              `json:"temperature,omitempty"`
	TopP                float32                              `json:"top_p,omitempty"`
	N                   int                                  `json:"n,omitempty"`
	Stream              bool                                 `json:"stream,omitempty"`
	Stop                []string                             `json:"stop,omitempty"`
	PresencePenalty     float32                              `json:"presence_penalty,omitempty"`
	ResponseFormat      *openai.ChatCompletionResponseFormat `json:"response_format,omitempty"`
	Seed                *int                                 `json:"seed,omitempty"`
	FrequencyPenalty    float32                              `json:"frequency_penalty,omitempty"`
	LogProbs            bool                                 `json:"logprobs,omitempty"`
	TopLogProbs         int                                  `json:"top_logprobs,omitempty"`
}

// Message represents a single message in the conversation.
type Message struct {
	// Role indicates the role of the message sender (e.g., "system", "user", "assistant").
	Role string `json:"role"`
	// Content contains the actual text of the message.
	Content string `json:"content"`
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func getTokens(text string) int {
	tkm, _ := tiktoken.GetEncoding("cl100k_base") // 使用 OpenAI 的分词模型
	tokens := tkm.Encode(text, nil, nil)
	return len(tokens)
}
