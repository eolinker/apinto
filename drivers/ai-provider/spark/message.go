package spark

type ClientRequest struct {
	Messages []*Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Response struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Choices []ResponseChoice `json:"choices"`
	Usage   Usage            `json:"usage"`
	Error   *Error           `json:"error"`
}

type ResponseChoice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

type Error struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param"`
	Code    interface{} `json:"code"`
}
