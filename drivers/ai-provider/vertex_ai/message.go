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
