package para_hmac

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

var signSort = []string{
	"Body",
	"X-App-Id",
	"X-Sequence-No",
	"X-Timestamp",
}

func sign(appId string, appKey string, timestamp string, sequenceNo string, body string) string {
	headerMap := make(map[string]string)
	headerMap["X-App-Id"] = appId
	headerMap["X-Sequence-No"] = sequenceNo
	headerMap["X-Timestamp"] = timestamp
	headerMap["Body"] = base64.StdEncoding.EncodeToString([]byte(body))
	builder := strings.Builder{}
	for _, key := range signSort {
		v, ok := headerMap[key]
		if !ok || v == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("%s=%s&", key, v))
	}
	builder.WriteString(appKey)
	h := sha256.New()
	h.Write([]byte(builder.String()))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
