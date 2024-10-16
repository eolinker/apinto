package vertex_ai

type ClientRequest struct {
	Contents []*Content `json:"contents"`
}

type Content struct {
	Parts []map[string]interface{} `json:"parts"`
	Role  string                   `json:"role"`
}

type Response struct {
	Candidates []Candidate `json:"candidates"`
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
