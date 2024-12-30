package openrouter

type ClientRequest struct {
	Messages []*Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Id      string           `json:"id"`
	Object  string           `json:"object"`
	Created int              `json:"created"`
	Model   string           `json:"model"`
	Choices []ResponseChoice `json:"choices"`
	Usage   Usage            `json:"usage"`
	Error   *Error           `json:"error"`
}

type ResponseChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// Error represents the error response from the provider.
// {"error":{"message":"Provider returned error","code":400,"metadata":{"raw":"{\n  \"error\": {\n    \"message\": \"Invalid value: 'yyy'. Supported values are: 'system', 'assistant', 'user', 'function', 'tool', and 'developer'.\",\n    \"type\": \"invalid_request_error\",\n    \"param\": \"messages[0].role\",\n    \"code\": \"invalid_value\"\n  }\n}","provider_name":"OpenAI"}},"user_id":"user_2nQFDPHnNOxsrry6JpmcPXFzfnC"}
type Error struct {
	Message  string `json:"message"`
	Code     int    `json:"code"`
	Metadata struct {
		Raw          string `json:"raw"`
		ProviderName string `json:"provider_name"`
	} `json:"metadata"`
}
