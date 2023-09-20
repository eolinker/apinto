package separator

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strconv"
	"strings"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var _ ICounter = (*FileCounter)(nil)

const defaultMultipartMemory = 32 << 20 // 32 MB

type FileCounter struct {
	typ      string
	split    string
	name     string
	max      int64
	splitLen int
}

func (f *FileCounter) Name() string {
	return f.name
}

func NewFileCounter(rule *CountRule) (*FileCounter, error) {
	var splitLen int
	if rule.SeparatorType == LengthCountType {
		var err error
		splitLen, err = strconv.Atoi(rule.Separator)
		if err != nil {
			splitLen = 1000
		}
	}
	return &FileCounter{name: rule.Key, split: rule.Separator, typ: rule.SeparatorType, max: rule.Max, splitLen: splitLen}, nil
}

func (f *FileCounter) Count(ctx http_service.IHttpContext) (int64, error) {
	raw, _ := ctx.Request().Body().RawBody()
	d, params, _ := mime.ParseMediaType(ctx.Request().ContentType())
	if !(d == "multipart/form-data") {
		return -1, fmt.Errorf("need content-type: multipart/form-data,now: %s", d)
	}
	boundary, ok := params["boundary"]
	if !ok {
		return -1, fmt.Errorf("missing boundary")
	}
	body := io.NopCloser(bytes.NewBuffer(raw))
	reader := multipart.NewReader(body, boundary)
	form, err := reader.ReadForm(defaultMultipartMemory)
	if err != nil {
		return -1, fmt.Errorf("parse form param err: %v", err)
	}

	switch f.typ {
	case LengthCountType:
		value := strings.Join(form.Value[f.name], "")
		l := len([]rune(value))
		if l%f.splitLen == 0 {
			return int64(l / f.splitLen), nil
		}
		return int64(l/f.splitLen + 1), nil
	case ArrayCountType:
		return 1, nil
	}
	return splitCount(strings.Join(form.Value[f.name], f.split), f.split), nil
}

func (f *FileCounter) Max() int64 {
	return f.max
}
