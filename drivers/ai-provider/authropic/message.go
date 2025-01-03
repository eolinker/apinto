package anthropic

type ClientRequest struct {
	Messages []*Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Response struct {
	Id         string    `json:"id"`
	Model      string    `json:"model"`
	Role       string    `json:"role"`
	Contents   []Content `json:"content"`
	StopReason string    `json:"stop_reason"`
	Usage      Usage     `json:"usage"`
	Error      Error     `json:"error"`
}
