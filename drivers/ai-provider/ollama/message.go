package ollama

import "time"

// ClientRequest represents the request payload sent to the Ollama API.
type ClientRequest struct {
	// Messages is a list of messages forming the conversation history.
	Messages []*Message `json:"messages"`
}

// Message represents a single message in the conversation.
type Message struct {
	// Role indicates the role of the message sender (e.g., "system", "user", "assistant").
	Role string `json:"role"`
	// Content contains the actual text of the message.
	Content string `json:"content"`
}

type Response struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            *Message  `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
	Error              string    `json:"error"`
}
