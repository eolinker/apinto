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
