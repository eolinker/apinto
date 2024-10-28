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

type Response struct {
	Id         string    `json:"id"`
	Model      string    `json:"model"`
	Role       string    `json:"role"`
	Contents   []Content `json:"content"`
	StopReason string    `json:"stop_reason"`
}
