package google

/*
*
返回示例

	{
	    "candidates": [
	        {
	            "content": {
	                "parts": [
	                    {
	                        "text": "Hello there! How can I help you today?\n"
	                    }
	                ],
	                "role": "model"
	            },
	            "finishReason": "STOP",
	            "avgLogprobs": -0.0011556809768080711
	        }
	    ],
	    "usageMetadata": {
	        "promptTokenCount": 2,
	        "candidatesTokenCount": 11,
	        "totalTokenCount": 13
	    },
	    "modelVersion": "gemini-1.5-flash-latest"
	}
*/
type ClientRequest struct {
	Contents []*Content `json:"contents"`
}

type Content struct {
	Parts []map[string]interface{} `json:"parts"`
	Role  string                   `json:"role"`
}

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type Response struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
	Error         Error         `json:"error"`
}

type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason"`
}

const (
	FinishReasonUnspecified     = "FINISH_REASON_UNSPECIFIED"
	FinishStop                  = "STOP"
	FinishMaxTokens             = "MAX_TOKENS"
	FinishSafety                = "SAFETY"
	FinishRecitation            = "RECITATION"
	FinishLanguage              = "LANGUAGE"
	FinishOther                 = "OTHER"
	FinishBlocklist             = "BLOCKLIST"
	FinishProhibitedContent     = "PROHIBITED_CONTENT"
	FinishSPII                  = "SPII"
	FinishMalformedFunctionCall = "MALFORMED_FUNCTION_CALL"
)
