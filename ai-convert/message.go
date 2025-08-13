package ai_convert

import (
	tiktoken "github.com/pkoukk/tiktoken-go"
	openai "github.com/sashabaranov/go-openai"
)

// Request 定义客户端统一输入请求格式
type Request struct {
	Model    string                         `json:"model"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
	// MaxTokens The maximum number of tokens that can be generated in the chat completion.
	// This value can be used to control costs for text generated via API.
	// This value is now deprecated in favor of max_completion_tokens, and is not compatible with o1 series models.
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat-create-max_tokens
	MaxTokens int `json:"max_tokens,omitempty"`
	// MaxCompletionTokens An upper bound for the number of tokens that can be generated for a completion,
	// including visible output tokens and reasoning tokens https://platform.openai.com/docs/guides/reasoning
	MaxCompletionTokens int      `json:"max_completion_tokens,omitempty"`
	Temperature         float32  `json:"temperature,omitempty"`
	TopP                float32  `json:"top_p,omitempty"`
	N                   int      `json:"n,omitempty"`
	Stream              bool     `json:"stream,omitempty"`
	Stop                []string `json:"stop,omitempty"`
	PresencePenalty     float32  `json:"presence_penalty,omitempty"`
}

// Response 定义客户端统一输出响应格式
type Response struct {
	*openai.ChatCompletionResponse `json:"response,omitempty"`
	Error                          *Error `json:"error,omitempty"`
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

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func getTokens(text string, model string) int {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		tkm, _ = tiktoken.GetEncoding(tiktoken.MODEL_CL100K_BASE) // 使用 OpenAI 的分词模型
	}

	tokens := tkm.Encode(text, nil, nil)
	return len(tokens)
}
