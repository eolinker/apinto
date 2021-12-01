package aksk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	http_service "github.com/eolinker/eosc/http-service"

	"github.com/eolinker/goku/auth"
)

const dateHeader = "x-gateway-date"

//buildToSign 构建待加密的签名所需字符串
func buildToSign(ctx http_service.IHttpContext, encType string, signedHeaders []string) string {
	toSign := strings.Builder{}
	toSign.WriteString(encType + "\n")
	dh := ctx.Request().Header().GetHeader(dateHeader)
	toSign.WriteString(dh + "\n")

	cr := buildHexCanonicalRequest(ctx, signedHeaders)
	toSign.WriteString(strings.ToLower(cr))
	return toSign.String()
}

//buildHexCanonicalRequest 构建规范消息头
func buildHexCanonicalRequest(ctx http_service.IHttpContext, signedHeaders []string) string {
	cr := strings.Builder{}

	cr.WriteString(strings.ToUpper(ctx.Request().Method()) + "\n")
	cr.WriteString(buildPath(ctx.Request().URI().Path()) + "\n")
	cr.WriteString(ctx.Request().URI().RawQuery() + "\n")

	for _, header := range signedHeaders {
		if strings.ToLower(header) == "host" {
			cr.WriteString(buildHeaders(header, ctx.Request().Header().Host()) + "\n")
			continue
		}
		v := ctx.Request().Header().GetHeader(header)
		cr.WriteString(buildHeaders(header, v) + "\n")
	}
	cr.WriteString("\n")
	cr.WriteString(strings.Join(signedHeaders, ";") + "\n")
	body, _ := ctx.Request().Body().RawBody()
	cr.WriteString(hexEncode(body))

	return hexEncode([]byte(cr.String()))
}

func buildPath(path string) string {
	return strings.TrimSuffix(path, "/") + "/"
}

func buildHeaders(hk, hv string) string {
	return fmt.Sprintf("%s:%s", hk, strings.TrimSpace(hv))
}

func hexEncode(data []byte) string {
	sha := sha256.New()
	sha.Write(data)
	return hex.EncodeToString(sha.Sum(nil))
}

func hmaxBySHA256(secretKey, toSign string) string {
	// 创建对应的sha256哈希加密算法
	hm := hmac.New(sha256.New, []byte(secretKey))
	//写入加密数据
	hm.Write([]byte(toSign))
	return hex.EncodeToString(hm.Sum(nil))
}

func parseAuthorization(ctx http_service.IHttpContext) (encType string, accessKey string, signHeaders []string, signature string, err error) {
	authStr := ctx.Request().Header().GetHeader(auth.Authorization)

	infos := strings.Split(authStr, ",")
	if len(infos) < 3 {
		err = errors.New("[ak/sk_auth] error authorization")
		return
	}
	encType, accessKey, err = parseAccessKey(infos[0])
	if err != nil {
		return
	}
	signHeaders, err = parseSignHeaders(infos[1])
	if err != nil {
		return
	}
	signature, err = parseSignature(infos[2])
	if err != nil {
		return
	}
	return
}

func parseAccessKey(info string) (string, string, error) {
	info = strings.TrimSpace(info)
	akInfos := strings.Split(info, " ")
	encType := ""
	accessKey := ""
	if len(akInfos) < 1 {
		return "", "", errors.New("[ak/sk_auth] error access key")
	} else if len(akInfos) == 1 {
		accessKey = strings.Replace(akInfos[0], "Access=", "", 1)
	} else if len(akInfos) == 2 {
		encType = akInfos[0]
		accessKey = strings.Replace(akInfos[1], "Access=", "", 1)
	}
	return encType, accessKey, nil
}

func parseSignHeaders(info string) ([]string, error) {
	info = strings.Replace(strings.TrimSpace(info), "SignedHeaders=", "", 1)
	headers := strings.Split(strings.ToLower(info), ";")
	return headers, nil
}

func parseSignature(info string) (string, error) {
	info = strings.Replace(strings.TrimSpace(info), "Signature=", "", 1)
	return info, nil
}
