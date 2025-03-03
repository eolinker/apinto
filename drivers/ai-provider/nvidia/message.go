package nvidia

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
}

// ErrorResponse represents an error response from the OpenAI API.
type ErrorResponse struct {
	Status int    `json:"status"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
	Title  string `json:"title"`
}

type ResponseChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
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
