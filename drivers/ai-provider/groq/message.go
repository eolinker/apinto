package groq

/*
*
返回正确示例

	{
	    "id": "chatcmpl-d615c27e-8f08-44f4-a0f1-241f4cb5731e",
	    "object": "chat.completion",
	    "created": 1735527988,
	    "model": "llama3-8b-8192",
	    "choices": [
	        {
	            "index": 0,
	            "message": {
	                "role": "assistant",
	                "content": "response content!"
	            },
	            "logprobs": null,
	            "finish_reason": "stop"
	        }
	    ],
	    "usage": {
	        "queue_time": 0.017568127,
	        "prompt_tokens": 18,
	        "prompt_time": 0.00188563,
	        "completion_tokens": 638,
	        "completion_time": 0.531666667,
	        "total_tokens": 656,
	        "total_time": 0.533552297
	    },
	    "system_fingerprint": "fp_179b0f92c9",
	    "x_groq": {
	        "id": "req_01jgarez1gea59fnyaxeq8m463"
	    }
	}

返回失败示例

	{
	    "error": {
	        "message": "'messages.0' : discriminator property 'role' has invalid value",
	        "type": "invalid_request_error"
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

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type Response struct {
	Id      string           `json:"id"`
	Object  string           `json:"object"`
	Created int              `json:"created"`
	Model   string           `json:"model"`
	Choices []ResponseChoice `json:"choices"`
	Usage   Usage            `json:"usage"`
	Error   Error            `json:"error"`
}

type ResponseChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}
