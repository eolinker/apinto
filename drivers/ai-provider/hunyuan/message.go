package hunyuan

type ClientRequest struct {
	Messages []*Message `json:"Messages"`
}

type Message struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

type Response struct {
	Response *ResponseInfo `json:"Response"`
}

type ResponseInfo struct {
	RequestId string `json:"RequestId"`
	Error     struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	} `json:"Error"`
	Note    string   `json:"Note"`
	Choices []Choice `json:"Choices"`
	Created int      `json:"Created"`
	Id      string   `json:"Id"`
	Usage   Usage    `json:"Usage"`
}

type Choice struct {
	FinishReason string  `json:"FinishReason"`
	Message      Message `json:"Message"`
}

type Usage struct {
	PromptTokens     int `json:"PromptTokens"`
	CompletionTokens int `json:"CompletionTokens"`
	TotalTokens      int `json:"TotalTokens"`
}
