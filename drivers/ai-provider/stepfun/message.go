package stepfun

/*
*
成功响应示例

	{
			"id": "5fd3d838464f358e4c33962e15092196.27dfabef4e49b8f8edd60e4074f8fe01",
			"object": "chat.completion",
			"created": 1735523297,
			"model": "step-1-8k",
			"choices": [
					{
							"index": 0,
							"message": {
									"role": "assistant",
									"content": ""
							},
							"finish_reason": "stop"
					}
			],
			"usage": {
					"prompt_tokens": 12,
					"completion_tokens": 745,
					"total_tokens": 757
			}
	}

响应失败示例

	{
	    "error": {
	        "message": "invalid msg role: assistant1",
	        "type": "request_params_invalid"
	    }
	}
*/
type ClientRequest struct {
	Messages []*Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Id      string           `json:"id"`
	Object  string           `json:"object"`
	Created int              `json:"created"`
	Model   string           `json:"model"`
	Choices []ResponseChoice `json:"choices"`
	Usage   Usage            `json:"usage"`
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
