package json_error

import (
	"encoding/json"
)

type JsonError struct {
	err   string
	errCN string
	code  string
}

func (e *JsonError) Error() string {
	result := map[string]string{
		"status": "err",
		"msg":    e.err,
		"code":   e.code,
	}
	if e.errCN != "" {
		result["msg_zh"] = e.errCN
	}
	resultByte, _ := json.Marshal(result)
	return string(resultByte)
}

func NewJsonError(errInfo string, errCN string, code string) *JsonError {
	return &JsonError{err: errInfo, errCN: errCN, code: code}
}
