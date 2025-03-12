package bedrock

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

type Response struct {
	Output     Output `json:"output"`
	StopReason string `json:"stopReason"`
}

type Output struct {
	Message *Message `json:"message"`
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

//
//// ConvertBedrockToOpenAI 通用转换方法
//func ConvertBedrockToOpenAI(requestId string, model string, bedrockResp BedrockResponse) openai.ChatCompletionResponse {
//	// 提取文本内容
//	textContent := ""
//	if len(bedrockResp.Output.Message.Content) > 0 {
//		textContent = bedrockResp.Output.Message.Content[0].Text
//	}
//	//const (
//	//	FinishReasonStop          FinishReason = "stop"
//	//	FinishReasonLength        FinishReason = "length"
//	//	FinishReasonFunctionCall  FinishReason = "function_call"
//	//	FinishReasonToolCalls     FinishReason = "tool_calls"
//	//	FinishReasonContentFilter FinishReason = "content_filter"
//	//	FinishReasonNull          FinishReason = "null"
//	//)
//	stopReason := openai.FinishReasonStop
//	//end_turn | tool_use | max_tokens | stop_sequence | guardrail_intervened | content_filtered
//	switch bedrockResp.StopReason {
//	case "max_tokens":
//		stopReason = openai.FinishReasonLength
//	case "content_filtered":
//		stopReason = openai.FinishReasonContentFilter
//	}
//	openai.FinishReason(bedrockResp.StopReason)
//	return openai.ChatCompletionResponse{
//		ID:      requestId,
//		Object:  "",
//		Created: 0, // 这里可以替换为实际时间戳
//		Model:   model,
//		Choices: []openai.ChatCompletionChoice{
//			{
//
//				FinishReason: stopReason,
//			},
//		},
//		Usage: struct {
//			PromptTokens     int `json:"prompt_tokens"`
//			CompletionTokens int `json:"completion_tokens"`
//			TotalTokens      int `json:"total_tokens"`
//		}{
//			PromptTokens:     bedrockResp.Usage.InputTokens,
//			CompletionTokens: bedrockResp.Usage.OutputTokens,
//			TotalTokens:      bedrockResp.Usage.TotalTokens,
//		},
//	}
//}
