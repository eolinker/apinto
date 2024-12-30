package minimax

/*
*
返回示例

	{
	    "id": "03c13c50d5d579fc940572bb6fdf6f0a",
	    "choices": [
	        {
	            "finish_reason": "stop",
	            "index": 0,
	            "message": {
	                "content": "I’m a language model, so I can answer questions about language, or about the world.",
	                "role": "assistant",
	                "name": "MM智能助理",
	                "audio_content": ""
	            }
	        }
	    ],
	    "created": 1735526736,
	    "model": "abab6.5s-chat",
	    "object": "chat.completion",
	    "usage": {
	        "total_tokens": 94,
	        "total_characters": 0,
	        "prompt_tokens": 75,
	        "completion_tokens": 19
	    },
	    "input_sensitive": false,
	    "output_sensitive": false,
	    "input_sensitive_type": 0,
	    "output_sensitive_type": 0,
	    "output_sensitive_int": 0,
	    "base_resp": {
	        "status_code": 0,
	        "status_msg": ""
	    }
	}
*/
type ClientRequest struct {
	Messages []*Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name"`
}

type Response struct {
	Id       string           `json:"id"`
	Object   string           `json:"object"`
	Created  int              `json:"created"`
	Model    string           `json:"model"`
	Choices  []ResponseChoice `json:"choices"`
	Usage    Usage            `json:"usage"`
	BaseResp BaseResp         `json:"base_resp"`
}

type BaseResp struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

type ResponseChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	TotalTokens      int `json:"total_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalCharacters  int `json:"total_characters"`
}

type CompletionTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}
