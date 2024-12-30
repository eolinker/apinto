package tongyi

type ClientRequest struct {
	Messages []*Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Id                string           `json:"id"`
	Object            string           `json:"object"`
	Created           int              `json:"created"`
	Model             string           `json:"model"`
	SystemFingerprint string           `json:"system_fingerprint"`
	Choices           []ResponseChoice `json:"choices"`
	Usage             Usage            `json:"usage"`
	Error             Error            `json:"error"`
}

type ResponseChoice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// Error represents an error response from the API.
// {"error":{"message":"Incorrect API key provided. ","type":"invalid_request_error","param":null,"code":"invalid_api_key"},"request_id":"2f09bd59-55cc-9f19-b785-6326783e9b05"}
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}
