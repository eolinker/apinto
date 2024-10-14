package wenxin

type ClientRequest struct {
	Messages []*Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Id           string `json:"id"`
	Object       string `json:"object"`
	Created      int    `json:"created"`
	Result       string `json:"result"`
	FinishReason string `json:"finish_reason"`
	ErrorCode    int    `json:"error_code"`
	ErrorMsg     string `json:"error_msg"`
}
