package google

import (
	"bufio"
	"bytes"
	"github.com/eolinker/apinto/encoder"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	"strings"
)

func StreamBodyParse(ctx http_service.IHttpContext, body []byte) []byte {
	encoding := ctx.Response().Headers().Get("content-encoding")
	target := body
	if encoding != "utf-8" && encoding != "" {
		tmp, err := encoder.ToUTF8(encoding, body)
		if err != nil {
			log.Errorf("convert to utf-8 error: %v, body: %s", err, string(body))
			return body
		}
		target = tmp
	}
	builder := strings.Builder{}
	scanner := bufio.NewScanner(bytes.NewReader(target))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimPrefix(line, "data:")
		if line == "" || strings.Trim(line, " ") == "[DONE]" {
			continue
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	return []byte(builder.String())
}
