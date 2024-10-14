package ai_provider

type ClientRequest struct {
	Messages []*Message `json:"messages"`
}

type ClientResponse struct {
	Message      Message `json:"message,omitempty"`
	FinishReason string  `json:"finish_reason,omitempty"`
	Code         int     `json:"code"`
	Error        string  `json:"error"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
