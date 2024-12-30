package cohere

type ClientRequest struct {
	Messages []*RequestMessage `json:"messages"`
}

type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Id           string      `json:"id"`
	Message      interface{} `json:"message"` // string or ResponseMessage
	Usage        Usage       `json:"usage"`
	FinishReason string      `json:"finish_reason"`
}

type ResponseContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ResponseMessage struct {
	Role    string            `json:"role"`
	Content []ResponseContent `json:"content"`
}

type Tokens struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Usage struct {
	BilledUnits Tokens `json:"billed_units"`
	Tokens      Tokens `json:"tokens"`
}
