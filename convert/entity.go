package convert

const (
	FinishReasonUnspecified     = "FINISH_REASON_UNSPECIFIED"
	FinishStop                  = "STOP"
	FinishMaxTokens             = "MAX_TOKENS"
	FinishSafety                = "SAFETY"
	FinishRecitation            = "RECITATION"
	FinishLanguage              = "LANGUAGE"
	FinishOther                 = "OTHER"
	FinishBlockList             = "BLOCKLIST"
	FinishProhibitedContent     = "PROHIBITED_CONTENT"
	FinishSPII                  = "SPII"
	FinishMalformedFunctionCall = "MALFORMED_FUNCTION_CALL"
)

type ClientRequest struct {
	Messages []*Message `json:"messages"`
	Stream   bool       `json:"stream"`
}

type ClientResponse struct {
	Message      *Message `json:"message,omitempty"`
	FinishReason string   `json:"finish_reason"`
	Code         int      `json:"code"`
	Error        string   `json:"error"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
