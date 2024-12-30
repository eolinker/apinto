package zhipuai

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

// Usage provides information about the token counts for the request and response.
// {"completion_tokens":217,"prompt_tokens":31,"total_tokens":248}
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
// {"error":{"code":"1002","message":"Authorization Token非法，请确认Authorization Token正确传递。"}}
type Error struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
