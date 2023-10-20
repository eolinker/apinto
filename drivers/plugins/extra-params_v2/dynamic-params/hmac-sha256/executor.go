package hmac_sha256

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/eolinker/eosc/log"

	dynamic_params "github.com/eolinker/apinto/drivers/plugins/extra-params_v2/dynamic-params"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type executor struct {
	*dynamic_params.Param
	secretKey string
}

func NewExecutor(name string, value []string) *executor {
	secretKey := ""
	v := value
	if len(value) >= 1 {
		secretKey = value[0]
		v = v[1:]
	}
	return &executor{
		Param:     dynamic_params.NewParam(name, v),
		secretKey: secretKey,
	}
}

func (m *executor) Generate(ctx http_service.IHttpContext, contentType string, args ...interface{}) (interface{}, error) {
	result, err := m.Param.Generate(ctx, contentType, args...)
	if err != nil {
		return nil, err
	}
	v, ok := result.(string)
	if !ok {
		return nil, errors.New("hmac-sha256 value is not string")
	}
	value := hMacBySHA256(m.secretKey, v)
	if !strings.HasPrefix(m.Name(), "__") {
		value = strings.ToUpper(value)
	}
	log.DebugF("hmac-sha256 value before: %s,after: %s", v, value)
	return value, nil
}

func hMacBySHA256(secretKey, toSign string) string {
	// 创建对应的sha256哈希加密算法
	hm := hmac.New(sha256.New, []byte(secretKey))
	//写入加密数据
	hm.Write([]byte(toSign))
	return hex.EncodeToString(hm.Sum(nil))
}
